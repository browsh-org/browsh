#!/usr/bin/env bash
set -e

function_to_run=$1

export PROJECT_ROOT
export GORELEASER_VERSION=1.10.2

PROJECT_ROOT=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

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

pushd "$PROJECT_ROOT" || _panic
"$function_to_run" "$@"
popd || _panic
