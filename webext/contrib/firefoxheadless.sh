#!/usr/bin/env bash

if [[ "$1" = "kill" ]]; then
	pkill --full 'firefox.*headless.*profile'
	sleep 1
	if [[ "$CI" == "true" ]]; then
		pkill -9 firefox || true
	fi
else
	FIREFOX_BIN=${FIREFOX:-firefox}
	"$FIREFOX_BIN" --headless --marionette "$@"
fi
