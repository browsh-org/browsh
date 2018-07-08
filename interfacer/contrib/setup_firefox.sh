#!/bin/bash

set -ex

if [ ! -f .travis.yml ]; then
  PROJECT_ROOT=$(git rev-parse --show-toplevel)
else
  PROJECT_ROOT=.
fi

line=$(cat $PROJECT_ROOT/.travis.yml | grep 'firefox: "')
version=$(echo $line | grep -o '".*"' | cut -d " " -f 1 | sed 's/"//g')

# Firefox is needed both for testing in Travis and embedding in the Docker
# image used by the Browsh as a Service platform. So we need to be able to
# give a specific and consistent version pin.
FIREFOX_VERSION=$version

mkdir -p $HOME/bin
pushd $HOME/bin
curl -L -o firefox.tar.bz2 https://ftp.mozilla.org/pub/firefox/releases/$FIREFOX_VERSION/linux-x86_64/en-US/firefox-$FIREFOX_VERSION.tar.bz2
bzip2 -d firefox.tar.bz2
tar xf firefox.tar
popd

firefox --version
