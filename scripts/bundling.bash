#!/usr/bin/env bash

export XPI_PATH="$PROJECT_ROOT"/interfacer/src/browsh/browsh.xpi
export XPI_SOURCE_DIR=$PROJECT_ROOT/webext/dist/web-ext-artifacts
export NODE_BIN=$PROJECT_ROOT/webext/node_modules/.bin
MDN_USER="user:13243312:78"

function versioned_xpi_file() {
	echo "$XPI_SOURCE_DIR/browsh-$(browsh_version).xpi"
}

# You'll want to use this with `go run ./cmd/browsh --debug --firefox.use-existing`
function build_webextension_watch() {
	pushd "$PROJECT_ROOT"/webext/dist || _panic
	"$NODE_BIN"/web-ext run \
		--firefox ../contrib/firefoxheadless.sh \
		--verbose
	popd || _panic
}

function build_webextension_production() {
	local version && version=$(browsh_version)

	cd "$PROJECT_ROOT"/webext && "$NODE_BIN"/webpack
	cd "$PROJECT_ROOT"/webext/dist && rm ./*.map
	if [ -f core ]; then
		# Is this a core dump for some failed process?
		rm core
	fi
	ls -alh .
	"$NODE_BIN"/web-ext build --overwrite-dest
	ls -alh web-ext-artifacts

	webextension_sign
	local source_file && source_file=$(versioned_xpi_file)

	echo "Bundling $source_file to $XPI_PATH"
	cp -f "$source_file" "$XPI_PATH"

	echo "Making extra copy for Goreleaser to put in Github release:"
	local goreleaser_pwd="$PROJECT_ROOT"/interfacer/
	cp -a "$source_file" "$goreleaser_pwd"
	ls -alh "$goreleaser_pwd"
}

# It is possible to use unsigned webextensions in Firefox but it requires that Firefox
# uses problematically insecure config. I know it's a hassle having to jump through all
# these signing hoops, but I think it's better to use a standard Firefox configuration.
# Moving away from the webextension alltogether is another story, but something I'm still
# thinking about.
#
# NB: There can only be one canonical XPI for each semantic version.
#
# shellcheck disable=2120
function webextension_sign() {
	local use_existing=$1
	if [ "$use_existing" == "" ]; then
		"$NODE_BIN"/web-ext sign --api-key "$MDN_USER" --api-secret "$MDN_KEY"
		_rename_built_xpi
	else
		echo "Skipping signing, downloading existing webextension"
		local base="https://github.com/browsh-org/browsh/releases/download"
		curl -L \
			-o "$(versioned_xpi_file)" \
			"$base/v$LATEST_TAGGED_VERSION/browsh-$LATEST_TAGGED_VERSION.xpi"
	fi
}

function _rename_built_xpi() {
	pushd "$XPI_SOURCE_DIR" || _panic
	local xpi_file
	xpi_file="$(
		find ./*.xpi \
			-printf "%T@ %f\n" |
			sort |
			cut -d' ' -f2 |
			tail -n1
	)"
	cp -a "$xpi_file" "$(versioned_xpi_file)"
	popd || _panic
}

function bundle_production_webextension() {
	local version && version=$(browsh_version)
	local base='https://github.com/browsh-org/browsh/releases/download'
	local release_url="$base/v$version/browsh-$version-an.fx.xpi"
	echo "Downloading webextension from: $release_url"
	local size && size=$(wc -c <"$XPI_PATH")
	curl -L -o "$XPI_PATH" "$release_url"
	if [ "$size" -lt 500 ]; then
		_panic "Problem downloading latest webextension XPI"
	fi
}
