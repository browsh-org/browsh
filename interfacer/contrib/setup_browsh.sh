#!/bin/bash

# I suspect this will be mostly used by automated CI.
# For example; in creating the Browsh Docker image. We don't actually build
# Browsh in the Dockerfile because that would require signing the webextension
# again, which can't be done as only one canonical release of a webextension is
# allowed by MDN per semantic version. It's actually quite good to not have to
# repeat the build process (after having done so in Travis after successfully
# passing tests). So we simply just download the already built binary :)

PROJECT_ROOT=$(git rev-parse --show-toplevel)
if [ $? -eq 0 ]; then
  manifest=$PROJECT_ROOT/webext/manifest.json
else
  manifest=./manifest.json
fi

line=$(cat $manifest | grep '"version"')
version=$(echo $line | grep -o '".*"' | cut -d " " -f 2 | sed 's/"//g')

base='https://github.com/browsh-org/browsh/releases/download'
release_url="$base/browsh-$version/browsh-linux-amd64-$version"

curl -L -o browsh $release_url
chmod a+x browsh
