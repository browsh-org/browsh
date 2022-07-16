#!/usr/bin/env bash

export WEBEXTENSION_GO=$PROJECT_ROOT/interfacer/src/browsh/webextension.go
export GOBINDATA_VERSION=3.23.0

# Convert the web extension into binary data that can be compiled into a
# cross-platform Go binary.
function xpi_to_bin() {
	local xpi_file=$1
	local bin_file=$2

	if ! command -v go-bindata &>/dev/null; then
		echo "Installing \`go-bindata'..."
		go install github.com/kevinburke/go-bindata/go-bindata@v"$GOBINDATA_VERSION"
		go-bindata -version
	fi

	go-bindata \
		-nocompress \
		-prefix "$PROJECT_ROOT/interfacer" \
		-pkg browsh \
		-o "$bin_file" \
		"$xpi_file"

	ls -alh "$WEBEXTENSION_GO"
}

function build_webextension() {
	local NODE_BIN=$PROJECT_ROOT/webext/node_modules/.bin

	cd "$PROJECT_ROOT"/webext && "$NODE_BIN"/webpack
	cd "$PROJECT_ROOT"/webext/dist && rm ./*.map
	if [ -f core ]; then
		# Is this a core dump for some failed process?
		rm core
	fi
	ls -alh .
	"$NODE_BIN"/web-ext build --overwrite-dest
	ls -alh web-ext-artifacts

	version=$("$PROJECT_ROOT"/ctl.sh browsh_version)

	local xpi_file=browsh-$version-an+fx.xpi
	local zip_file=browsh-$version.zip
	local source_dir=$PROJECT_ROOT/webext/dist/web-ext-artifacts

	if [ "$BROWSH_ENV" == "RELEASE" ]; then
		# The signed version. There can only be one canonical XPI for each semantic
		# version.
		source_file=$source_dir/$xpi_file
		bundle_file=$PROJECT_ROOT/interfacer/browsh.xpi
		"$NODE_BIN"/web-ext sign --api-key "$MDN_USER" --api-secret "$MDN_KEY"
	else
		# TODO: This doesn't currently work with the Marionettte `tempAddon`
		# installation method. Just use `web-ext run` and Browsh's `use-existing-ff`
		# flag - which is better anyway as it auto-reloads the extension when files
		# change. NB: If you fix this, don't forget to change the filename loaded
		# by `Asset()` in `main.go`.
		# In development/testing, we want to be able to bundle the webextension
		# frequently without having to resort to version bumps.
		source_file=$source_dir/$zip_file
		bundle_file=$source_dir/browsh.zip
	fi

	cp -f "$source_file" "$bundle_file"
	echo "Bundling $source_file to $WEBEXTENSION_GO using internal path $bundle_file"
	xpi2bin "$bundle_file" "$WEBEXTENSION_GO"
}

function bundle_production_webextension() {
	local version && version=$(browsh_version)
	local base='https://github.com/browsh-org/browsh/releases/download'
	local release_url="$base/v$version/browsh-$version-an.fx.xpi"
	local xpi_file=$PROJECT_ROOT/interfacer/src/browsh/browsh.xpi
	curl -L -o "$xpi_file" "$release_url"

	xpi2bin "$xpi_file" "$WEBEXTENSION_GO"
}
