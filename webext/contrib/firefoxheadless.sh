#!/bin/bash
FIREFOX_BIN=${FIREFOX:-firefox}
$FIREFOX_BIN --headless "$@"
