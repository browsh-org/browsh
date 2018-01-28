#!/bin/bash
set -e

# Install `dep` the current defacto dependency for Golang
GOLANG_DEP_VERSION=0.3.2
dep_url=https://github.com/golang/dep/releases/download/v$GOLANG_DEP_VERSION/dep-linux-amd64
curl -L -o $GOPATH/bin/dep $dep_url
chmod +x $GOPATH/bin/dep

