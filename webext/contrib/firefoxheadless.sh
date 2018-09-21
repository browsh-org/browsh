#!/usr/bin/env bash

if [[ "$1" = "kill" ]]; then
  kill $(ps aux|grep headless|grep 'profile /tmp'| tr -s ' ' | cut -d ' ' -f2)
else
  FIREFOX_BIN=${FIREFOX:-firefox}
  $FIREFOX_BIN --headless "$@"
fi
