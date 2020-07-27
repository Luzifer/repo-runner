module github.com/repo-runner/repo-runner/cmd/repo-runner

go 1.14

replace github.com/repo-runner/repo-runner => ../../

require (
	github.com/Luzifer/go_helpers/v2 v2.10.0
	github.com/Luzifer/rconfig/v2 v2.2.1
	github.com/ejholmes/hookshot v0.0.0-20170729003051-9585af7fb64a
	github.com/fsouza/go-dockerclient v1.6.5
	github.com/gofrs/uuid v3.3.0+incompatible
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/websocket v1.4.2
	github.com/hpcloud/tail v1.0.0
	github.com/repo-runner/repo-runner v0.0.0-00010101000000-000000000000
)
