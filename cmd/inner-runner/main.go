package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/Luzifer/rconfig"
	reporunner "github.com/Luzifer/repo-runner"
	homedir "github.com/mitchellh/go-homedir"
)

var (
	cfg = struct {
		VersionAndExit bool `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	version = "dev"
)

func init() {
	if err := rconfig.Parse(&cfg); err != nil {
		log.Fatalf("Unable to parse commandline options: %s", err)
	}

	if cfg.VersionAndExit {
		fmt.Printf("inner-runner %s\n", version)
		os.Exit(0)
	}
}

func main() {
	netrcLocation, err := homedir.Expand("~/.netrc")

	log.Printf("[INFO] Setting access token for HTTPs clone")
	netrcContent := fmt.Sprintf("machine github.com\nlogin auth\npassword %s", os.Getenv("GITHUB_TOKEN"))
	if err := ioutil.WriteFile(netrcLocation, []byte(netrcContent), 0600); err != nil {
		log.Fatalf("[FATA] Unable to write ~/.netrc: %s", err)
	}

	log.Printf("[INFO] Checking out repository to /src")
	if err := execute("", "/usr/bin/git", "clone", os.Getenv("CLONE_URL"), "/src"); err != nil {
		log.Fatalf("[FATA] Could not clone repository: %s", err)
	}

	log.Printf("[INFO] Checking out rev %s in repository", os.Getenv("REVISION"))
	if err := execute("/src", "/usr/bin/git", "reset", "--hard", os.Getenv("REVISION")); err != nil {
		log.Fatalf("[FATA] Could not check out revision: %s", err)
	}

	runnerFile, err := reporunner.LoadFromFile("/src/.repo-runner.yaml")
	if err != nil {
		log.Fatalf("[FATA] Could not load runner-configuration: %s", err)
	}

	for _, cmd := range runnerFile.Commands {
		if err := execute("/src", "/bin/sh", "-c", cmd); err != nil {
			log.Fatalf("[FATA] Command exitted non-zero, stopping now.")
		}
	}
}

func execute(dir, name string, args ...string) error {
	log.Printf("[INFO] Exec: %s %s", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = os.Environ()

	return cmd.Run()
}
