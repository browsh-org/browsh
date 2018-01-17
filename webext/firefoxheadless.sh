#!/bin/bash
firefox-beta --headless "$@" &
pid=$!
trap "kill ${pid}; exit 1" INT
wait
