module github.com/repo-runner/repo-runner/cmd/repo-runner

go 1.14

replace github.com/repo-runner/repo-runner => ../../

require (
	github.com/Luzifer/go_helpers/v2 v2.10.0
	github.com/Luzifer/rconfig/v2 v2.2.1
	github.com/Microsoft/hcsshim v0.8.9 // indirect
	github.com/containerd/containerd v1.3.6 // indirect
	github.com/containerd/continuity v0.0.0-20200710164510-efbc4488d8fe // indirect
	github.com/ejholmes/hookshot v0.0.0-20170729003051-9585af7fb64a
	github.com/fsouza/go-dockerclient v1.6.5
	github.com/gofrs/uuid v3.3.0+incompatible
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/websocket v1.4.2
	github.com/hpcloud/tail v1.0.0
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/repo-runner/repo-runner v0.16.0
	github.com/sirupsen/logrus v1.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208 // indirect
	golang.org/x/sys v0.0.0-20200724161237-0e2f3a69832c // indirect
	google.golang.org/genproto v0.0.0-20200726014623-da3ae01ef02d // indirect
	google.golang.org/grpc v1.30.0 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/validator.v2 v2.0.0-20200605151824-2b28d334fa05 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
)
