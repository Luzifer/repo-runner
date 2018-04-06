ci: publish

qa:
	go get gopkg.in/alecthomas/gometalinter.v1
	gometalinter.v1 --vendored-linters --install
	gometalinter.v1 \
		-D gotype -D errcheck -D gas -D gocyclo \
		--sort path --sort line --deadline 1m --cyclo-over 15 \
		. ./cmd/repo-runner ./cmd/inner-runner

publish:
	curl -sSLo golang.sh https://raw.githubusercontent.com/Luzifer/github-publish/master/golang.sh
	bash golang.sh
