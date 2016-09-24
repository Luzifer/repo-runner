package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/Luzifer/go_helpers/env"
	"github.com/Luzifer/rconfig"
	"github.com/Luzifer/repo-runner"
	"github.com/ejholmes/hookshot"
	docker "github.com/fsouza/go-dockerclient"
	uuid "github.com/satori/go.uuid"
)

var (
	cfg = struct {
		DefaultEnv     []string      `flag:"default-env,e" default:"" description:"Environment variables to set when starting the container"`
		DockerSocket   string        `flag:"docker-sock" default:"unix:///var/run/docker.sock" description:"Docker socket / tcp address"`
		GithubToken    string        `flag:"github-token" env:"GITHUB_TOKEN" default:"" description:"Personal Access Token to fetch config from private Repos"`
		Listen         string        `flag:"listen" default:":3000" description:"IP/Port to listen on"`
		LogDir         string        `flag:"log-dir" default:"./logs/" description:"Where to write build logs?"`
		MaxBuildTime   time.Duration `flag:"max-build-time" default:"1h" description:"Maximum time the build may run"`
		RequireSecret  string        `flag:"require-secret" default:"" description:"(Optional) Require a secret when receiving the hookshot"`
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

	r := hookshot.NewRouter()

	if cfg.RequireSecret == "" {
		r.Handle("ping", pingHandler{})
		r.Handle("push", pushHandler{})
	} else {
		r.Handle("ping", hookshot.Authorize(pingHandler{}, cfg.RequireSecret))
		r.Handle("push", hookshot.Authorize(pushHandler{}, cfg.RequireSecret))
	}

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
	if err := setGithubBuildStatus(ctx, payload.Repository.FullName, payload.After, "pending",
		"Build started with ID "+logID); err != nil {
		log.Printf("[ERRO] (%s | %.7s) Could not set Github build status: %s",
			payload.Repository.FullName, payload.After, err)
	}

	buildLog := bytes.NewBuffer([]byte{})
	defer func() {
		logPath := path.Join(cfg.LogDir, logID+".txt")

		if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
			log.Printf("[ERRO] (%s | %.7s) Could not ensure log dir: %s",
				payload.Repository.FullName, payload.After, err)
			return
		}

		if err := ioutil.WriteFile(logPath, buildLog.Bytes(), 0644); err != nil {
			log.Printf("[ERRO] (%s | %.7s) Could not write log file: %s",
				payload.Repository.FullName, payload.After, err)
			return
		}

		log.Printf("[INFO] (%s | %.7s) Build log was written to %s",
			payload.Repository.FullName, payload.After, logPath)
	}()

	runnerFile, err := repo_runner.LoadFromGithub(payload.Repository.FullName, cfg.GithubToken)
	if err != nil {
		log.Printf("[ERRO] (%s | %.7s) Could not fetch runner file: %s",
			payload.Repository.FullName, payload.After, err)
		return
	}

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

	envVars := env.MapToList(envMap)

	dockerRepo, dockerTag := docker.ParseRepositoryTag(runnerFile.Image)
	auth, authAvailable := dockerAuth.Configs[strings.SplitN(dockerRepo, "/", 2)[0]]
	if !authAvailable {
		auth = docker.AuthConfiguration{}
	}

	log.Printf("[INFO] (%s | %.7s) Refreshing docker image '%s'",
		payload.Repository.FullName, payload.After, runnerFile.Image)
	if err := dockerClient.PullImage(docker.PullImageOptions{
		Repository:   dockerRepo,
		Tag:          dockerTag,
		OutputStream: buildLog,
		Context:      ctx,
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
			Volumes: map[string]struct{}{},
			Mounts:  []docker.Mount{}, //TODO: Fill me
		},
	})
	if err != nil {
		log.Printf("[ERRO] (%s | %.7s) Could not create container: %s",
			payload.Repository.FullName, payload.After, err)
		return
	}

	log.Printf("[INFO] (%s | %.7s) Starting build with container '%s'",
		payload.Repository.FullName, payload.After, container.Name)
	if err := dockerClient.StartContainer(container.ID, &docker.HostConfig{}); err != nil {
		log.Printf("[ERRO] (%s | %.7s) Starting container failed: %s",
			payload.Repository.FullName, payload.After, err)
		return
	}

	log.Printf("[INFO] (%s | %.7s) Attaching to container logs",
		payload.Repository.FullName, payload.After)
	cw, err := dockerClient.AttachToContainerNonBlocking(docker.AttachToContainerOptions{
		Container:    container.ID,
		OutputStream: buildLog,
		ErrorStream:  buildLog,
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
				if err := setGithubBuildStatus(ctx, payload.Repository.FullName, payload.After, "success",
					fmt.Sprintf("Build with ID %s exited with status 0", logID)); err != nil {
					log.Printf("[ERRO] (%s | %.7s) Could not set Github build status: %s",
						payload.Repository.FullName, payload.After, err)
				}
			} else {
				if err := setGithubBuildStatus(ctx, payload.Repository.FullName, payload.After, "failure",
					fmt.Sprintf("Build with ID %s exited with status %d", logID, ct.State.ExitCode)); err != nil {
					log.Printf("[ERRO] (%s | %.7s) Could not set Github build status: %s",
						payload.Repository.FullName, payload.After, err)
				}
			}

			keepWaiting = false
		}
	}
}
