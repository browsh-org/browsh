#!/bin/env bash

function _panic() {
	local message=$1
	echo >&2 "$message"
	exit 1
}
