default: bundle_assets

bundle_assets: fetch_deps
	go generate

fetch_deps:
	bash ./buildAssets.sh
