#!/bin/bash

# I suspect this will be mostly used by automated CI.
# For example; in creating the Browsh Docker image. We don't actually build
# Browsh in the Dockerfile because that would require signing the webextension
# again, which can't be done as only one canonical release of a webextension is
# allowed by MDN per semantic version. It's actually quite good to not have to
# repeat the build process (having done so in Travis after successfully
# passing tests). So we simply just download the already built binary.

if [ ! -f manifest.json ]; then
  PROJECT_ROOT=$(git rev-parse --show-toplevel)/webext
else
  PROJECT_ROOT=.
fi
manifest=$PROJECT_ROOT/manifest.json

line=$(cat $manifest | grep '"version"')
version=$(echo $line | grep -o '".*"' | cut -d " " -f 2 | sed 's/"//g')

base='https://github.com/tombh/texttop/releases/download'
release_url="$base/v$version/browsh_${version}_linux_amd64"

curl -L -o browsh $release_url
chmod a+x browsh
