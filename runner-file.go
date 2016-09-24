package reporunner

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	yaml "gopkg.in/yaml.v2"
)

const (
	defaultRunnerFileLocation = ".repo-runner.yaml"
	defaultFetchTimeout       = 5 * time.Second
)

// RunnerFile contains the instructions what to run when executing build for the specific repo
type RunnerFile struct {
	Image       string            `yaml:"image"`
	Commands    []string          `yaml:"commands"`
	Environment map[string]string `yaml:"environment"`
}

type ghFileResponse struct {
	Content string `json:"content"`
}

// LoadFromGithub uses the Github API to fetch the RunnerFile from the repo before pulling the repository
func LoadFromGithub(repo, token string) (*RunnerFile, error) {
	// https://developer.github.com/v3/repos/contents/#get-contents
	// GET /repos/:owner/:repo/contents/:path

	u := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s",
		repo, defaultRunnerFileLocation)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	if token != "" {
		req.SetBasicAuth("auth", token)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultFetchTimeout)
	defer cancel()

	res, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Could not fetch %s file due to HTTP %d error.",
			defaultRunnerFileLocation, res.StatusCode)
	}

	ghr := ghFileResponse{}
	if err = json.NewDecoder(res.Body).Decode(&ghr); err != nil {
		return nil, err
	}

	yamlData, err := base64.StdEncoding.DecodeString(ghr.Content)
	if err != nil {
		return nil, err
	}

	rf := &RunnerFile{}
	return rf, yaml.Unmarshal(yamlData, rf)
}

// LoadFromFile loads a local RunnerFile
func LoadFromFile(filename string) (*RunnerFile, error) {
	if _, err := os.Stat(filename); err != nil {
		return nil, err
	}

	yamlData, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	rf := &RunnerFile{}
	return rf, yaml.Unmarshal(yamlData, rf)
}
