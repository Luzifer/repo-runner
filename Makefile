VERSION=$(shell git describe --tags --exact-match)

ci: vet publish

vet:
	go vet ./cmd/repo-runner ./cmd/inner-runner

ifneq ($(strip $(VERSION)),)
publish:
	VERSION=$(VERSION) sh -e publish.sh
else
publish:
	true
endif

