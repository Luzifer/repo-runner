VERSION := $(shell git describe --tags --exact-match)

ci: qa publish

qa:
	go get gopkg.in/alecthomas/gometalinter.v1
	go vet . ./cmd/repo-runner ./cmd/inner-runner
	go test . ./cmd/repo-runner ./cmd/inner-runner
	gometalinter.v1 \
		-D gotype -D errcheck \
		--sort path --sort line --deadline 1m --cyclo-over 15 \
		. ./cmd/repo-runner ./cmd/inner-runner

ifneq ($(strip $(VERSION)),)
publish:
	VERSION=$(VERSION) sh -e publish.sh
else
publish:
	true
endif

