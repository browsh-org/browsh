#!/bin/bash
set -e

GOLANG_VERSION=1.9.3
GOLANG_DEP_VERSION=0.3.2

export GOPATH=$HOME/gopath
export PATH=$PATH:/$HOME/bin:$GOPATH/bin
dep_url=https://github.com/golang/dep/releases/download/v$GOLANG_DEP_VERSION/dep-linux-amd64
golang_archive=go$GOLANG_VERSION.linux-amd64.tar.gz
golang_url=https://dl.google.com/go/$golang_archive

mkdir -p $HOME/bin
mkdir -p $GOPATH/bin

# Install Golang
curl -L -o $golang_archive $golang_url
tar -C $HOME/bin -xzf $golang_archive

# Install `dep` the current defacto dependency for Golang
curl -L -o $GOPATH/bin/dep $dep_url
chmod +x $GOPATH/bin/dep

