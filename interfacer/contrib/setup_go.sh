#!/bin/bash
set -e

# Install `dep` the current defacto dependency for Golang
GOLANG_DEP_VERSION=0.3.2
dep_url=https://github.com/golang/dep/releases/download/v$GOLANG_DEP_VERSION/dep-linux-amd64

if [ -z $GOPATH ]; then
  echo '$GOPATH is not set, aborting'
  exit 1
fi

curl -L -o $GOPATH/bin/dep $dep_url
chmod +x $GOPATH/bin/dep

