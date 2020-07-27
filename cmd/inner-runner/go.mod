module github.com/repo-runner/repo-runner/cmd/inner-runner

go 1.14

replace github.com/repo-runner/repo-runner => ../../

require (
	github.com/Luzifer/rconfig/v2 v2.2.1
	github.com/mitchellh/go-homedir v1.1.0
	github.com/repo-runner/repo-runner v0.16.0
	github.com/spf13/pflag v1.0.5 // indirect
	gopkg.in/validator.v2 v2.0.0-20200605151824-2b28d334fa05 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
)
