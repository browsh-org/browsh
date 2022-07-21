#!/usr/bin/env bash

export BROWSH_VERSION
export LATEST_TAGGED_VERSION

function _goreleaser_production() {
	if ! command -v goreleaser &>/dev/null; then
		echo "Installing \`goreleaser'..."
		go install github.com/goreleaser/goreleaser@v"$GORELEASER_VERSION"
	fi
	pushd "$PROJECT_ROOT"/interfacer || _panic
	_export_versions
	[ "$BROWSH_VERSION" = "" ] && _panic "BROWSH_VERSION unset (goreleaser needs it)"
	goreleaser release \
		--config "$PROJECT_ROOT"/goreleaser.yml \
		--rm-dist
	popd || _panic
}

function _export_versions() {
	BROWSH_VERSION=$(_parse_browsh_version)
	LATEST_TAGGED_VERSION=$(
		git tag --sort=v:refname --list 'v*.*.*' | tail -n1 | sed -e "s/^v//"
	)
}

function _parse_browsh_version() {
	local version_file=$PROJECT_ROOT/interfacer/src/browsh/version.go
	local line && line=$(grep 'browshVersion' <"$version_file")
	local version && version=$(echo "$line" | grep -o '".*"' | sed 's/"//g')
	echo -n "$version"
}

function _is_new_version() {
	_export_versions
	[ "$BROWSH_VERSION" = "" ] && _panic "BROWSH_VERSION unset"
	[ "$LATEST_TAGGED_VERSION" = "" ] && _panic "LATEST_TAGGED_VERSION unset"
	[[ "$BROWSH_VERSION" != "$LATEST_TAGGED_VERSION" ]]
}

function _tag_on_version_change() {
	_export_versions
	echo_versions

	if ! _is_new_version; then
		echo "Not tagging as there's no new version."
		exit 0
	fi

	git tag v"$BROWSH_VERSION"
	git show v"$BROWSH_VERSION" --quiet
	git config --global user.email "ci@github.com"
	git config --global user.name "Github Actions"
	git add --all
	git reset --hard v"$BROWSH_VERSION"
}

function echo_versions() {
	_export_versions
	echo "Browsh binary version: $BROWSH_VERSION"
	echo "Git latest tag: $LATEST_TAGGED_VERSION"
}

function browsh_version() {
	_export_versions
	echo -n "$BROWSH_VERSION"
}

function github_actions_output_version_status() {
	local status="false"
	if _is_new_version; then
		status="true"
	fi
	echo "::set-output name=is_new_version::$status"
}

function webext_build_release() {
	pushd "$PROJECT_ROOT"/webext || _panic
	build_webextension_production
	popd || _panic
}

function update_browsh_website_with_new_version() {
	_export_versions
	local remote="git@github.com:browsh-org/www.brow.sh.git"
	pushd /tmp || _panic
	git clone "$remote"
	cd www.brow.sh || _panic
	echo "latest_version: $BROWSH_VERSION" >_data/browsh.yml
	git add _data/browsh.yml
	git commit -m "Github Actions: updated Browsh version to $BROWSH_VERSION"
	git push "$remote"
	popd || _panic
}

function update_homebrew_tap_with_new_version() {
	_export_versions
	local remote="git@github.com:browsh-org/homebrew-browsh.git"
	pushd /tmp || _panic
	git clone "$remote"
	cd homebrew-browsh || _panic
	cp -f "$PROJECT_ROOT"/interfacer/dist/browsh.rb browsh.rb
	git add browsh.rb
	git commit -m "Github Actions: updated to $BROWSH_VERSION"
	git push "$remote"
	popd || _panic
}

function goreleaser_local_only() {
	pushd "$PROJECT_ROOT"/interfacer || _panic
	goreleaser release \
		--config "$PROJECT_ROOT"/goreleaser.yml \
		--snapshot \
		--rm-dist
	popd || _panic
}

function build_browsh_binary() {
	# Requires $path argument because it's used in the Dockerfile where the GOROOT is
	# outside .git/
	local path=$1
	pushd "$path" || _panic
	local webextension="src/browsh/browsh.xpi"
	[ ! -f "$webextension" ] && _panic "browsh.xpi not present"
	md5sum "$webextension"
	go build ./cmd/browsh
	echo "Freshly built \`browsh' version: $(./browsh --version 2>&1)"
	popd || _panic
}

function release() {
	[ "$(git rev-parse --abbrev-ref HEAD)" != "master" ] && _panic "Not releasing unless on the master branch"
	webext_build_release
	build_browsh_binary "$PROJECT_ROOT"/interfacer
	_tag_on_version_change
	_goreleaser_production
}
