#!/bin/bash
set -ex

pushd dist
curl -sL -o upx.txz https://github.com/upx/upx/releases/download/v3.94/upx-3.94-amd64_linux.tar.xz
tar -xvf upx.txz
cp upx-3.94-amd64_linux/upx .
rm -rf upx-3.94-amd64_linux
./upx */*
popd
