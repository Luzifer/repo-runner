package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
)

type pushPayload struct {
	Ref        string `json:"ref"`
	Before     string `json:"before"`
	After      string `json:"after"`
	Repository struct {
		FullName string `json:"full_name"`
		Private  bool   `json:"private"`
		CloneURL string `json:"clone_url"`
	} `json:"repository"`
}

func (p pushPayload) String() string {
	buf := bytes.NewBuffer([]byte{})
	b64buf := base64.NewEncoder(base64.StdEncoding, buf)
	json.NewEncoder(b64buf).Encode(p)
	return buf.String()
}
