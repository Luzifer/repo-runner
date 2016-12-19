package main

//go:generate go-bindata -pkg $GOPACKAGE -o assets.go assets/

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/Luzifer/go_helpers/env"
	"github.com/Luzifer/rconfig"
	"github.com/ejholmes/hookshot"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/gorilla/mux"
	reporunner "github.com/repo-runner/repo-runner"
	uuid "github.com/satori/go.uuid"
)

var (
	cfg = struct {
		BaseURL        string        `flag:"base-url" env:"BASE_URL" default:"http://127.0.0.1:3000" description:"URL this is reachable at (for build status URLs)"`
		DefaultEnv     []string      `flag:"default-env,e" default:"" description:"Environment variables to set when starting the container"`
		DefaultMount   []string      `flag:"default-mount,v" default:"" description:"Mountpoints to be forced into the container"`
		DockerSocket   string        `flag:"docker-sock" default:"unix:///var/run/docker.sock" description:"Docker socket / tcp address"`
		GithubToken    string        `flag:"github-token" env:"GITHUB_TOKEN" default:"" description:"Personal Access Token to fetch config from private Repos"`
		Listen         string        `flag:"listen" default:":3000" description:"IP/Port to listen on"`
		LogDir         string        `flag:"log-dir" default:"./logs/" description:"Where to write build logs?"`
		MaxBuildTime   time.Duration `flag:"max-build-time" default:"1h" description:"Maximum time the build may run"`
		Privileged     bool          `flag:"privileged" default:"false" description:"Run container privileged"`
		RequireSecret  string        `flag:"require-secret" env:"HOOK_SECRET" default:"" description:"(Optional) Require a secret when receiving the hookshot"`
		VersionAndExit bool          `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	version = "dev"

	dockerClient *docker.Client
	dockerAuth   *docker.AuthConfigurations
)

func init() {
	if err := rconfig.Parse(&cfg); err != nil {
		log.Fatalf("Unable to parse commandline options: %s", err)
	}

	if cfg.VersionAndExit {
		fmt.Printf("repo-runner %s\n", version)
		os.Exit(0)
	}
}

func main() {
	var err error
	dockerClient, err = docker.NewClient(cfg.DockerSocket)
	if err != nil {
		log.Fatalf("[FATA] Could not connect to docker socket: %s", err)
	}

	dockerAuth, err = docker.NewAuthConfigurationsFromDockerCfg()
	if err != nil {
		log.Printf("[WARN] Could not load docker auth configuration")
	}

	hr := hookshot.NewRouter()

	if cfg.RequireSecret == "" {
		hr.Handle("ping", pingHandler{})
		hr.Handle("push", pushHandler{})
	} else {
		hr.Handle("ping", hookshot.Authorize(pingHandler{}, cfg.RequireSecret))
		hr.Handle("push", hookshot.Authorize(pushHandler{}, cfg.RequireSecret))
	}

	r := mux.NewRouter()
	r.Handle("/hook", hr)
	registerLogHandlers(r.PathPrefix("/log").Subrouter())

	http.ListenAndServe(cfg.Listen, r)
}

type pingHandler struct{}

func (p pingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte(`Pong`))
}

type pushHandler struct{}

func (p pushHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload := pushPayload{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Did not understand your JSON body.", http.StatusBadRequest)
		return
	}

	go startJob(payload)

	w.WriteHeader(http.StatusNoContent)
}

func startJob(payload pushPayload) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.MaxBuildTime)
	defer cancel()

	logID := uuid.NewV4().String()
	buildStatus := &githubBuildStatus{
		Repo:        payload.Repository.FullName,
		SHA:         payload.After,
		State:       "pending",
		Description: "Build started with ID " + logID,
		TargetURL:   strings.TrimRight(cfg.BaseURL, "/") + "/log/" + logID,
	}

	runnerFile, err := reporunner.LoadFromGithub(payload.Repository.FullName, cfg.GithubToken)
	if err != nil {
		log.Printf("[FATA] (%s | %.7s) Could not fetch runner file: %s",
			payload.Repository.FullName, payload.After, err)
		return
	}

	rex, err := regexp.Compile(runnerFile.AllowBuild)
	if err != nil {
		log.Printf("[FATA] (%s | %.7s) Invalid regular expression in allow_build: %s",
			payload.Repository.FullName, payload.After, err)
		return
	}

	if !rex.MatchString(payload.Ref) {
		log.Printf("[INFO] (%s | %.7s) Stopping because ref does not match allow_build ('%s'): %s",
			payload.Repository.FullName, payload.After, runnerFile.AllowBuild, payload.Ref)
		return
	}

	if err = buildStatus.Set(ctx); err != nil {
		log.Printf("[ERRO] (%s | %.7s) Could not set Github build status: %s",
			payload.Repository.FullName, payload.After, err)
	}

	buildStatus.State = "error"
	buildStatus.Description = "An unknown build error occurred"
	defer func() { buildStatus.Set(context.Background()) }()

	logPath := path.Join(cfg.LogDir, logID+".jsonl")

	if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
		log.Printf("[FATA] (%s | %.7s) Could not ensure log dir: %s",
			payload.Repository.FullName, payload.After, err)
		return
	}
	buildLog, err := os.Create(logPath)
	if err != nil {
		log.Printf("[FATA] (%s | %.7s) Could not open log file for writing: %s",
			payload.Repository.FullName, payload.After, err)
		return
	}
	buildLogMeta := newLogWriter("meta", buildLog)
	buildLogStdOut := newLogWriter("stdout", buildLog)
	buildLogStdErr := newLogWriter("stderr", buildLog)

	buildLogMeta.MetaMessage(logTypeMetaStart, payload.String())
	buildLogMeta.MetaMessage(logTypeMetaRepoName, payload.Repository.FullName)
	buildLogMeta.MetaMessage(logTypeMetaRepoURL, payload.Repository.HTMLURL)

	defer func() {
		log.Printf("[INFO] (%s | %.7s) Build log was written to %s",
			payload.Repository.FullName, payload.After, logPath)
		buildLog.Close()
	}()

	envMap := env.ListToMap(cfg.DefaultEnv)
	if runnerFile.Environment != nil {
		for k, v := range runnerFile.Environment {
			envMap[k] = v
		}
	}

	envMap["CLONE_URL"] = payload.Repository.CloneURL
	envMap["REVISION"] = payload.After
	envMap["PAYLOAD"] = payload.String()
	envMap["GITHUB_TOKEN"] = cfg.GithubToken
	envMap["CHECKOUT_DIR"] = runnerFile.CheckoutDir

	envVars := env.MapToList(envMap)

	mounts, volumes, binds := parseMounts(cfg.DefaultMount)

	dockerRepo, dockerTag := docker.ParseRepositoryTag(runnerFile.Image)
	auth, authAvailable := dockerAuth.Configs[strings.SplitN(dockerRepo, "/", 2)[0]]
	if !authAvailable {
		auth = docker.AuthConfiguration{}
	}

	log.Printf("[INFO] (%s | %.7s) Refreshing docker image '%s'",
		payload.Repository.FullName, payload.After, runnerFile.Image)
	if err = dockerClient.PullImage(docker.PullImageOptions{
		Repository: dockerRepo,
		Tag:        dockerTag,
		Context:    ctx,
	}, auth); err != nil {
		log.Printf("[ERRO] (%s | %.7s) Could not refresh docker image '%s': %s",
			payload.Repository.FullName, payload.After, runnerFile.Image, err)
		return
	}

	log.Printf("[INFO] (%s | %.7s) Creating container",
		payload.Repository.FullName, payload.After)
	container, err := dockerClient.CreateContainer(docker.CreateContainerOptions{
		Name: logID,
		Config: &docker.Config{
			Image:   runnerFile.Image,
			Env:     envVars,
			Volumes: volumes,
			Mounts:  mounts,
		},
		HostConfig: &docker.HostConfig{
			Binds:      binds,
			Privileged: cfg.Privileged,
		},
	})
	if err != nil {
		log.Printf("[ERRO] (%s | %.7s) Could not create container: %s",
			payload.Repository.FullName, payload.After, err)
		return
	}

	log.Printf("[INFO] (%s | %.7s) Starting build with container '%s'",
		payload.Repository.FullName, payload.After, container.Name)
	if err = dockerClient.StartContainer(container.ID, &docker.HostConfig{}); err != nil {
		log.Printf("[ERRO] (%s | %.7s) Starting container failed: %s",
			payload.Repository.FullName, payload.After, err)
		return
	}

	log.Printf("[INFO] (%s | %.7s) Attaching to container logs",
		payload.Repository.FullName, payload.After)
	cw, err := dockerClient.AttachToContainerNonBlocking(docker.AttachToContainerOptions{
		Container:    container.ID,
		OutputStream: buildLogStdOut,
		ErrorStream:  buildLogStdErr,
		Logs:         true,
		Stream:       true,
		Stdout:       true,
		Stderr:       true,
	})
	if err != nil {
		log.Printf("[ERRO] (%s | %.7s) Could not attach to container logs: %s",
			payload.Repository.FullName, payload.After, err)
		return
	}

	doneChan := make(chan error)
	go func() { doneChan <- cw.Wait() }()

	keepWaiting := true

	for keepWaiting {
		select {
		case <-ctx.Done():
			if err := dockerClient.StopContainer(container.ID, 30); err != nil {
				log.Printf("[ERRO] (%s | %.7s) Stopping container failed: %s",
					payload.Repository.FullName, payload.After, err)
			}
		case <-doneChan:
			ct, err := dockerClient.InspectContainer(container.ID)
			if err != nil {
				log.Printf("[ERRO] (%s | %.7s) Could not fetch exit status of the container: %s",
					payload.Repository.FullName, payload.After, err)
			}

			log.Printf("[INFO] (%s | %.7s) Work is done or time is over. Build exited with status %d",
				payload.Repository.FullName, payload.After, ct.State.ExitCode)
			if err := dockerClient.RemoveContainer(docker.RemoveContainerOptions{
				ID:            container.ID,
				RemoveVolumes: true,
				Force:         true,
			}); err != nil {
				log.Printf("[ERRO] (%s | %.7s) Removing container failed: %s",
					payload.Repository.FullName, payload.After, err)
			}

			if ct.State.ExitCode == 0 {
				buildStatus.State = "success"
				buildStatus.Description = fmt.Sprintf("Build with ID %s exited with status 0", logID)
			} else {
				buildStatus.State = "failure"
				buildStatus.Description = fmt.Sprintf("Build with ID %s exited with status %d", logID, ct.State.ExitCode)
			}

			buildLogMeta.MetaMessage(logTypeMetaFinished, buildStatus.State)

			keepWaiting = false
		}
	}
}

func parseMounts(mountIn []string) (mounts []docker.Mount, volumes map[string]struct{}, binds []string) {
	volumes = make(map[string]struct{})
	for _, m := range mountIn {
		if len(m) == 0 {
			continue
		}

		parts := strings.Split(m, ":")
		if len(parts) != 2 && len(parts) != 3 {
			log.Printf("[ERRO] Invalid default mount: %s", m)
			continue
		}

		if stat, err := os.Stat(parts[0]); err == nil && !stat.IsDir() {
			binds = append(binds, m)
			continue
		}

		mo := docker.Mount{
			Source:      parts[0],
			Destination: parts[1],
		}

		if len(parts) == 3 {
			mo.RW = (parts[3] != "ro")
		}

		mounts = append(mounts, mo)
		volumes[mo.Destination] = struct{}{}
	}

	return
}
