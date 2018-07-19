#!/bin/bash

# I suspect this will be mostly used by automated CI.
# For example; in creating the Browsh Docker image. We don't actually build
# Browsh in the Dockerfile because that would require signing the webextension
# again, which can't be done as only one canonical release of a webextension is
# allowed by MDN per semantic version. It's actually quite good to not have to
# repeat the build process (having done so in Travis after successfully
# passing tests). So we simply just download the already built binary.

set -e

PROJECT_ROOT=$(git rev-parse --show-toplevel)
version=$($PROJECT_ROOT/contrib/get_browsh_version.sh)

base='https://github.com/browsh-org/browsh/releases/download'
release_url="$base/v$version/browsh_${version}_linux_amd64"

curl -L -o browsh $release_url
chmod a+x browsh
