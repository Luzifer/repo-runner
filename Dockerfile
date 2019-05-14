FROM golang:alpine

LABEL maintainer Knut Ahlers <knut@ahlers.me>

ENV CGO_ENABLED=0

COPY . /go/src/github.com/repo-runner/repo-runner
WORKDIR /go/src/github.com/repo-runner/repo-runner

RUN set -ex \
 && apk add --no-cache git ca-certificates \
 && go install -ldflags "-X main.version=$(git describe --tags || git rev-parse --short HEAD || echo dev)" github.com/repo-runner/repo-runner/cmd/repo-runner \
 && apk --no-cache del --purge git

EXPOSE 3000

VOLUME ["/var/run/docker.sock", "/root/.docker", "/var/log/repo-runner"]

ENTRYPOINT ["/go/bin/repo-runner", "--log-dir=/var/log/repo-runner/"]
CMD ["--"]
