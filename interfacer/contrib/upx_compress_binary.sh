#!/bin/bash
set -ex

pushd dist
upx */*
popd
