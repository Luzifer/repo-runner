FROM golang:alpine

MAINTAINER Knut Ahlers <knut@ahlers.me>

ADD . /go/src/github.com/Luzifer/repo-runner/cmd/inner-runner

RUN set -ex \
 && apk add --update git ca-certificates make \
 && go install -ldflags "-X main.version=$(git describe --tags || git rev-parse --short HEAD || echo dev)" github.com/Luzifer/repo-runner/cmd/inner-runner \
 && mv /go/bin/inner-runner /usr/local/bin/inner-runner \
 && rm -rf /go/src /go/bin /go/pkg

ENTRYPOINT ["/usr/local/bin/inner-runner"]
CMD ["--"]
