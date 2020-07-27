module github.com/repo-runner/repo-runner

go 1.14

replace (
	github.com/repo-runner/repo-runner/cmd/inner-runner => ./cmd/inner-runner
	github.com/repo-runner/repo-runner/cmd/repo-runner => ./cmd/repo-runner
)
