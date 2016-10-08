FROM golang:alpine

MAINTAINER Knut Ahlers <knut@ahlers.me>

ENV CGO_ENABLED=0

ADD . /go/src/github.com/Luzifer/repo-runner/cmd/repo-runner
WORKDIR /go/src/github.com/Luzifer/repo-runner/cmd/repo-runner

RUN set -ex \
 && apk add --update git ca-certificates \
 && go install -ldflags "-X main.version=$(git describe --tags || git rev-parse --short HEAD || echo dev)" \
 && apk del --purge git

EXPOSE 3000

VOLUME ["/var/run/docker.sock", "/root/.docker/config.json", "/var/log/repo-runner"]

ENTRYPOINT ["/go/bin/repo-runner", "--log-dir=/var/log/repo-runner/"]
CMD ["--"]
