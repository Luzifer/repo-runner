VERSION=$(shell git describe --tags)

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

