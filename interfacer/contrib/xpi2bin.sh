#!/bin/bash
set -e

INTERFACER_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && cd ../ && pwd )"

go-bindata -version
go-bindata \
  -nocompress \
  -prefix $INTERFACER_ROOT \
  -pkg browsh \
  -o $BIN_FILE \
  $XPI_FILE

ls -alh $INTERFACER_ROOT/src/browsh/webextension.go
echo "go-bindata exited with $(echo $?)"
