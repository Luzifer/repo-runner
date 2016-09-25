#!/bin/sh -e

if [ -z "${VERSION}" ]; then
	echo "Please set \$VERSION environment variable"
	exit 1
fi

if [ -z "${GITHUB_TOKEN}" ]; then
	echo "PLease set \$GITHUB_TOKEN environment variable"
	exit 1
fi

set -x

go get github.com/aktau/github-release
go get github.com/mitchellh/gox

github-release release --user Luzifer --repo repo-runner --tag ${VERSION} --name ${VERSION} || true

for binary in repo-runner inner-runner; do
	cd cmd/${binary}
	gox -ldflags="-X main.version=${VERSION}" -osarch="linux/386 linux/amd64 linux/arm"
	for file in ${binary}_*; do
		github-release upload --user Luzifer --repo repo-runner --tag ${VERSION} --name ${file} --file ${file}
	done
  cd -
done
