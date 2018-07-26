#!/bin/bash
set -e

INTERFACER_ROOT=$(readlink -m "$( cd "$(dirname "$0")" ; pwd -P )"/../)

go-bindata -version
go-bindata \
  -nocompress \
  -prefix $INTERFACER_ROOT \
  -pkg browsh \
  -o $BIN_FILE \
  $XPI_FILE

ls -alh $INTERFACER_ROOT/src/browsh/webextension.go
echo "go-bindata exited with $(echo $?)"
