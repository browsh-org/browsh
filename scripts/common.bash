#!/usr/bin/env bash

# shellcheck disable=2120
function _panic() {
	local message=$1
	echo >&2 "$message"
	exit 1
}

function _md5() {
	local path=$1
	md5sum "$path" | cut -d' ' -f1
}

function pushd() {
	# shellcheck disable=2119
	command pushd "$@" >/dev/null || _panic
}

function popd() {
	# shellcheck disable=2119
	command popd "$@" >/dev/null || _panic
}
