#!/bin/bash

FIREFOX_VERSION=58.0b16

mkdir -p $HOME/bin
pushd $HOME/bin
curl -L -o firefox.tar.bz2 https://ftp.mozilla.org/pub/firefox/releases/$FIREFOX_VERSION/linux-x86_64/en-US/firefox-$FIREFOX_VERSION.tar.bz2
apt-get -y install bzip2
bzip2 -d firefox.tar.bz2
tar xf firefox.tar
popd
