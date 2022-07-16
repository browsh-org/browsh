#!/usr/bin/env bash

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

	version=$(browsh_version)

	local source_file
	local source_dir=$PROJECT_ROOT/webext/dist/web-ext-artifacts
	local bundle_file=$PROJECT_ROOT/interfacer/src/browsh/browsh.xpi

	if [ "$BROWSH_ENV" == "RELEASE" ]; then
		# The signed version. There can only be one canonical XPI for each semantic
		# version.
		source_file="$source_dir/browsh-$version-an+fx.xpi"
		"$NODE_BIN"/web-ext sign --api-key "$MDN_USER" --api-secret "$MDN_KEY"
	else
		# TODO: This doesn't currently work with the Marionettte `tempAddon`
		# installation method. Just use `web-ext run` and Browsh's `use-existing-ff`
		# flag - which is better anyway as it auto-reloads the extension when files
		# change. NB: If you fix this, don't forget to change the filename loaded
		# by `Asset()` in `main.go`.
		# In development/testing, we want to be able to bundle the webextension
		# frequently without having to resort to version bumps.
		source_file="$source_dir/browsh-$version.zip"
	fi

	echo "Bundling $source_file to $bundle_file"
	cp -f "$source_file" "$bundle_file"
}

function bundle_production_webextension() {
	local version && version=$(browsh_version)
	local base='https://github.com/browsh-org/browsh/releases/download'
	local release_url="$base/v$version/browsh-$version-an.fx.xpi"
	local xpi_file=$PROJECT_ROOT/interfacer/src/browsh/browsh.xpi
	curl -L -o "$xpi_file" "$release_url"
}
