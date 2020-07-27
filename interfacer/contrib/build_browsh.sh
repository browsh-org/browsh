#!/usr/bin/env bash

# This is for building a production version of Browsh.
# To build Browsh during development see:
#   https://github.com/browsh-org/browsh#contributing

# This script depends on Golang, go-bindata and curl
# `go-bindata` can be easily installed with:
#   `go get -u gopkg.in/shuLhan/go-bindata.v3/...`

set -e

INTERFACER_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && cd ../ && pwd )"
cd $INTERFACER_ROOT

# Install the tool to convert the web extenstion file into a Go-compatible binary
go get -u gopkg.in/shuLhan/go-bindata.v3/...

# Get the current Browsh version, in order to find the corresponding web extension release
version_file=$INTERFACER_ROOT/src/browsh/version.go
line=$(grep 'browshVersion' < $version_file)
version=$(echo $line | grep -o '".*"' | sed 's/"//g')

# Build the URI for the webextension file
base='https://github.com/browsh-org/browsh/releases/download'
release_url="$base/v$version/browsh-${version}-an.fx.xpi"

xpi_file=$INTERFACER_ROOT/browsh.xpi
destination=$INTERFACER_ROOT/src/browsh/webextension.go

# Download the web extension
curl -L -o $xpi_file $release_url

# Convert the web extension into binary data that can be compiled into a
# cross-platform Go binary.
XPI_FILE=$xpi_file BIN_FILE=$destination \
  $INTERFACER_ROOT/contrib/xpi2bin.sh

# The actual build iteself
go build ./cmd/browsh
