#!/bin/env bash

export BROWSH_VERSION
export LATEST_TAGGED_VERSION

function _goreleaser_production() {
	if ! command -v goreleaser &>/dev/null; then
		echo "Installing \`goreleaser'..."
		go install github.com/goreleaser/goreleaser@v"$GORELEASER_VERSION"
	fi
	pushd "$PROJECT_ROOT"/interfacer/src || _panic
	goreleaser release
	popd || _panic
}

function _export_versions() {
	BROWSH_VERSION=$(_parse_browsh_version)
	LATEST_TAGGED_VERSION=$(
		git tag --sort=v:refname --list 'v*.*.*' | tail -n1 | sed -e "s/^v//"
	)
}

function _parse_browsh_version() {
	version_file=$PROJECT_ROOT/interfacer/src/browsh/version.go
	line=$(grep 'browshVersion' <"$version_file")
	version=$(echo "$line" | grep -o '".*"' | sed 's/"//g')
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
		echo "Not running release as there's no new version."
		exit 0
	fi

	git tag v"$BROWSH_VERSION"
	git show v"$BROWSH_VERSION" --quiet
	git config --global user.email "ci@github.com"
	git config --global user.name "Github Actions"
	# `/dev/null` needed to prevent Github token appearing in logs
	git push --tags --quiet https://"$GITHUB_TOKEN"@github.com/browsh-org/browsh >/dev/null 2>&1

	git reset --hard v"$BROWSH_VERSION"
}

function echo_versions() {
	_export_versions
	echo "Browsh Golang version: $BROWSH_VERSION"
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

function npm_build_release() {
	pushd "$PROJECT_ROOT"/webext || _panic
	BROWSH_ENV=RELEASE npm run build_webextension
	popd || _panic
}

function update_browsh_website_with_new_version() {
	pushd /tmp || _panic
	git clone https://github.com/browsh-org/www.brow.sh.git
	cd www.brow.sh || exit 1
	echo "latest_version: $BROWSH_VERSION" >_data/browsh.yml
	git add _data/browsh.yml
	git commit -m "Github Actions: updated Browsh version to $BROWSH_VERSION"
	# `/dev/null` needed to prevent Github token appearing in logs
	git push --quiet https://"$GITHUB_TOKEN"@github.com/browsh-org/www.brow.sh >/dev/null 2>&1
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

function release() {
	npm_build_release
	_tag_on_version_change
	_goreleaser_production
	update_browsh_website_with_new_version
}
