#!/bin/bash
set -euxo pipefail

css_deps=(
	npm/bootstrap@3.4.1/dist/css/bootstrap.min.css
)

js_deps=(
	npm/jquery@3.5.1
	npm/bootstrap@3.4.1/dist/js/bootstrap.min.js
	npm/jquery.scrollto@2.1.2
)

IFS=$','
curl -sSfLo ./assets/bundle.css "https://cdn.jsdelivr.net/combine/${css_deps[*]}"
curl -sSfLo ./assets/bundle.js "https://cdn.jsdelivr.net/combine/${js_deps[*]}"

# Bundling font
curl -sSfLo ./assets/font.zip "https://google-webfonts-helper.herokuapp.com/api/fonts/source-code-pro?download=zip&subsets=latin&variants=regular&formats=woff,woff2"
pushd assets
unzip font.zip
rm font.zip
popd
