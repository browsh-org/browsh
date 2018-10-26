#!/usr/bin/env bash

set -e

PROJECT_ROOT=$(git rev-parse --show-toplevel)
version_file=$PROJECT_ROOT/interfacer/src/browsh/version.go
line=$(grep 'browshVersion' < $version_file)
version=$(echo $line | grep -o '".*"' | sed 's/"//g')
echo -n $version
