#!/bin/env bash
set -e

function_to_run=$1

export PROJECT_ROOT && PROJECT_ROOT=$(git rev-parse --show-toplevel)
export GORELEASER_VERSION=1.10.2
export GOBINDATA_VERSION=3.23.0

function _includes_path {
	echo "$PROJECT_ROOT"/scripts
}

function _load_includes {
	for file in "$(_includes_path)"/*.bash; do
		# shellcheck disable=1090
		source "$file"
	done
}

_load_includes

if [[ $(type -t "$function_to_run") != function ]]; then
	echo "Subcommand: '$function_to_run' not found."
	exit 1
fi

shift
"$function_to_run" "$@"
