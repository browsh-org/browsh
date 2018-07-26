#!/bin/bash

# This is for building a production version of Browsh.
# To build Browsh during development see:
#   https://github.com/browsh-org/browsh#contributing

# This script depends on Golang and go-bindata
# `go get -u gopkg.in/shuLhan/go-bindata.v3/...`

set -e

INTERFACER_ROOT=$(readlink -m "$( cd "$(dirname "$0")" ; pwd -P )"/../)

version_file=$INTERFACER_ROOT/src/browsh/version.go
line=$(cat $version_file | grep 'browshVersion')
version=$(echo $line | grep -o '".*"' | sed 's/"//g')

base='https://github.com/browsh-org/browsh/releases/download'
release_url="$base/v$version/browsh-${version}-an.fx.xpi"

xpi_file=$INTERFACER_ROOT/browsh.xpi
destination=$INTERFACER_ROOT/src/browsh/webextension.go

curl -L -o $xpi_file $release_url

XPI_FILE=$xpi_file BIN_FILE=$destination \
  $INTERFACER_ROOT/contrib/xpi2bin.sh

cd $INTERFACER_ROOT
go build -o browsh src/main.go
