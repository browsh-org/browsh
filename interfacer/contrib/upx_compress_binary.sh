#!/bin/bash
set -ex
shopt -s extglob

pushd dist
upx !(@(freebsd*|openbsd*|darwin*|linux_arm64))/*
popd
