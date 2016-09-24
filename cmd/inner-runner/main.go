package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/Luzifer/rconfig"
	"github.com/Luzifer/repo-runner"
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
	log.Printf("[INFO] Checking out repository to /src")
	if err := execute("", "/usr/bin/git", "clone", os.Getenv("SSH_URL"), "/src"); err != nil {
		log.Fatalf("[FATA] Could not clone repository, stopping now.")
	}

	log.Printf("[INFO] Checking out rev %s in repository")
	if err := execute("/src", "/usr/bin/git", "reset", "--hard", os.Getenv("REVISION")); err != nil {
		log.Fatalf("[FATA] Could not check out revision, stopping now.")
	}

	runnerFile, err := repo_runner.LoadFromFile("/src/.repo-runner.yaml")
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
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = os.Environ()

	return cmd.Run()
}
