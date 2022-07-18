#!/usr/bin/env bash

function golang_lint_check() {
	pushd "$PROJECT_ROOT"/interfacer || _panic
	diff -u <(echo -n) <(gofmt -d ./)
	popd || _panic
}

function golang_lint_fix() {
	gofmt -w ./interfacer
}

function prettier_fix() {
	pushd "$PROJECT_ROOT"/webext || _panic
	prettier --write '{src,test}/**/*.js'
	popd || _panic
}

function parse_firefox_version_from_ci_config() {
	local line && line=$(grep 'firefox-version:' <"$PROJECT_ROOT"/.github/workflows/main.yml)
	local version && version=$(echo "$line" | tr -s ' ' | cut -d ' ' -f 3)
	[ "$version" = "" ] && _panic "Couldn't parse Firefox version"
	echo -n "$version"
}

function install_firefox() {
	local version && version=$(parse_firefox_version_from_ci_config)
	local destination=/tmp
	echo "Installing Firefox v$version to $destination..."
	mkdir -p "$destination"
	pushd "$destination" || _panic
	curl -L -o firefox.tar.bz2 \
		"https://ftp.mozilla.org/pub/firefox/releases/$version/linux-x86_64/en-US/firefox-$version.tar.bz2"
	bzip2 -d firefox.tar.bz2
	tar xf firefox.tar
	popd || _panic
}

function parse_golang_version_from_ci_config() {
	local line && line=$(grep 'go-version:' <"$PROJECT_ROOT"/.github/workflows/main.yml)
	local version && version=$(echo "$line" | tr -s ' ' | cut -d ' ' -f 3)
	[ "$version" = "" ] && _panic "Couldn't parse Golang version"
	echo -n "$version"
}

function install_golang() {
	local version && version=$(parse_golang_version_from_ci_config)
	[ "$GOPATH" = "" ] && _panic "GOPATH not set"
	[ "$GOROOT" = "" ] && _panic "GOROOT not set"
	echo "Installing Golang v$version... to $GOROOT"
	curl -L \
		-o go.tar.gz \
		https://dl.google.com/go/go"$version".linux-amd64.tar.gz
	mkdir -p "$GOPATH"/bin
	mkdir -p "$GOROOT"
	tar -C "$GOROOT/.." -xzf go.tar.gz
	go version
}

function build_browsh_binary() {
	local path=$1
	pushd "$path" || _panic
	local webextension="src/browsh/browsh.xpi"
	[ ! -f "$webextension" ] && _panic "browsh.xpi not present"
	md5sum "$webextension"
	go build ./cmd/browsh
	./browsh --version
	popd || _panic
}
