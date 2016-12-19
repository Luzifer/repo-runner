# Luzifer / repo-runner

`repo-runner` is a daemon executing docker containers to build repository pushes to Github repositories. It could also have been called "Yet another CI" as in the end it is a simple CI.

## Features

- Receive Github webhook events
- Spawn Docker containers in which the build is done
- Set Github build status for commits
- Stream logs to web interface

## Components

The repo runner consists of two parts: The `repo-runner` and the `inner-runner`

- `repo-runner` component is the main daemon to be run available to the web. It receives the push, fetches the configuration for that specific repo and starts the container.
- `inner-runner` is a small executable to be put into the build container. It receives some environment vars from the `repo-runner`, clones the repoistory into a configurable path and afterwards executes custom commands.

## Configuration

### repo-runner

All configuration for the daemon is done using commandline flags which are available including descriptions using `repo-runner --help`. The Github token and the hook secret can be set using ENV vars. (`GITHUB_TOKEN` and `HOOK_SECRET`)

The Github token is required as soon as one of the following features is to be used:

- Building of private repositories
- Setting the Github build status

It is a [personal access token](https://github.com/settings/tokens) and should have `repo` permission.

### Configuration file

The configuration file to be put into the repository is quite simple:

```yaml
---
image: "quay.io/luzifer/repo-runner-image"
checkout_dir: /go/src/github.com/Luzifer/repo-runner
allow_build: '^refs/heads/.*'

commands:
  - make ci

environment:
  CGO_ENABLED: 0
```

- `image` - The docker image to be executed
- `checkout_dir` - The directory to put sources to (default `/src`)
- `allow_build` - Filter what types of pushes to build (default `^refs/heads/.*`)
  - `^refs/heads/.*` = Build all branches but no tags
  - `^refs/tags/.*` = Build all tags but no branches
  - `^refs/heads/master$` = Build only master branch
  - `.*` = Build everything
- `commands` - Commands to be executed. They will get executed in the `checkout_dir`
- `environment` - Variables to be set when executing the build

## Using another build image

To exchange the build image is quite simple: You can use every image you can find out there. It just needs to follow some simple points:

- It gets these environment variables to work with:
  - `CLONE_URL` - The HTTPs clone URL of the repo
  - `REVISION` - The SHA of the commit which caused the push
  - `PAYLOAD` - Part of the hook event (JSON string, base64 encoded)
  - `GITHUB_TOKEN` - Passed through from the daemon
  - `CHECKOUT_DIR` - Directory the source should be cloned to
- It needs to take care of cloning the source
- It needs to take care of executing CI commands
- It must have an entrypoint or a command specified in the image
- It should end with `exit 0` if the build was successful or something other if not

Most easy way to accomplish this is to put the `inner-runner` into the image as it does all those tasks.
