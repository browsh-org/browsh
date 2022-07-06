#!/usr/bin/env bash

if [[ "$1" = "kill" ]]; then
  pids=$(ps aux|grep headless|grep 'profile '| tr -s ' ' | cut -d ' ' -f2)
  if [[ $pids =~ [^0-9] ]] ; then
    kill $pids
  fi
  if [[ "$CI" == "true" ]]; then
    pkill -9 firefox || true
  fi
else
  FIREFOX_BIN=${FIREFOX:-firefox}
  $FIREFOX_BIN --headless --marionette "$@"
fi
