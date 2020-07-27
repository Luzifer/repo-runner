#!/bin/bash
set -euxo pipefail

css_deps=(
	npm/bootstrap@4.5.0/dist/css/bootstrap.min.css
	npm/bootstrap-vue@2.15.0/dist/bootstrap-vue.min.css
	npm/bootswatch@4.5.0/dist/darkly/bootstrap.min.css
)

js_deps=(
	npm/vue@2.6.11
	npm/bootstrap-vue@2.15.0/dist/bootstrap-vue.min.js
)

IFS=$','
curl -sSfLo ./assets/bundle.css "https://cdn.jsdelivr.net/combine/${css_deps[*]}"
curl -sSfLo ./assets/bundle.js "https://cdn.jsdelivr.net/combine/${js_deps[*]}"

# Bundling font
here=$(pwd)
fontdir=$(mktemp -d)
trap "rm -rf ${fontdir}" EXIT

pushd ${fontdir}

# Source Code Pro Google Font
curl -sSfLo font.zip "https://google-webfonts-helper.herokuapp.com/api/fonts/source-code-pro?download=zip&subsets=latin&variants=regular&formats=woff,woff2"
unzip font.zip
cp source-code-pro* "${here}/assets/"

curl -sSfLo fa.zip "https://use.fontawesome.com/releases/v5.14.0/fontawesome-free-5.14.0-web.zip"
unzip fa.zip
cp fontawesome-free-*/sprites/solid.svg "${here}/assets/fa-solid.svg"

popd
