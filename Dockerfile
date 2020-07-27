FROM golang:alpine as builder

COPY . /go/src/github.com/repo-runner/repo-runner
WORKDIR /go/src/github.com/repo-runner/repo-runner

RUN set -ex \
 && apk add --update git \
 && go install \
      -ldflags "-X main.version=$(git describe --tags --always || echo dev)" \
      -mod=readonly


FROM alpine:latest

ENV LOG_DIR=/var/log/repo-runner/

LABEL maintainer "Knut Ahlers <knut@ahlers.me>"

RUN set -ex \
 && apk --no-cache add \
      ca-certificates

COPY --from=builder /go/bin/repo-runner /usr/local/bin/repo-runner

EXPOSE 3000
VOLUME ["/var/run/docker.sock", "/root/.docker", "/var/log/repo-runner"]

ENTRYPOINT ["/usr/local/bin/repo-runner"]
CMD ["--"]

# vim: set ft=Dockerfile:
