FROM golang:alpine

MAINTAINER Knut Ahlers <knut@ahlers.me>

ADD . /go/src/github.com/repo-runner/repo-runner/cmd/inner-runner

RUN set -ex \
 && apk add --update git ca-certificates make curl bash \
 && go install -ldflags "-X main.version=$(git describe --tags || git rev-parse --short HEAD || echo dev)" github.com/Luzifer/repo-runner/cmd/inner-runner \
 && mv /go/bin/inner-runner /usr/local/bin/inner-runner \
 && rm -rf /go/src /go/bin /go/pkg \
 && curl -sSL https://get.docker.com/builds/Linux/x86_64/docker-1.12.1.tgz | tar -xz -C /tmp \
 && mv /tmp/docker/* /usr/local/bin \
 && rm -rf /tmp/docker

ENTRYPOINT ["/usr/local/bin/inner-runner"]
CMD ["--"]
