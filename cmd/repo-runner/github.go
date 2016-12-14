package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type pushPayload struct {
	Ref        string `json:"ref"`
	Before     string `json:"before"`
	After      string `json:"after"`
	Repository struct {
		FullName string `json:"full_name"`
		Private  bool   `json:"private"`
		CloneURL string `json:"clone_url"`
		HTMLURL  string `json:"html_url"`
	} `json:"repository"`
}

func (p pushPayload) String() string {
	buf := bytes.NewBuffer([]byte{})
	b64buf := base64.NewEncoder(base64.StdEncoding, buf)
	json.NewEncoder(b64buf).Encode(p)
	return buf.String()
}

type githubBuildStatus struct {
	Repo, SHA, State, Description, TargetURL string
}

func (g githubBuildStatus) Set(ctx context.Context) error {
	// https://developer.github.com/v3/repos/statuses/#create-a-status
	// POST /repos/:owner/:repo/statuses/:sha

	if cfg.GithubToken == "" {
		return errors.New("Status can only get set when Gitub token is available")
	}

	buf := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(buf).Encode(struct {
		State       string `json:"state"`
		Description string `json:"description,omitempty"`
		Context     string `json:"context"`
		TargetURL   string `json:"target_url,omitempty"`
	}{
		State:       g.State,
		Description: g.Description,
		Context:     "continuous-integration/repo-runner",
		TargetURL:   g.TargetURL,
	}); err != nil {
		return err
	}

	u := fmt.Sprintf("https://api.github.com/repos/%s/statuses/%s", g.Repo, g.SHA)
	req, err := http.NewRequest("POST", u, buf)
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("auth", cfg.GithubToken)

	reqctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := http.DefaultClient.Do(req.WithContext(reqctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		return fmt.Errorf("Received unexpected status code: %d", res.StatusCode)
	}

	return nil
}
