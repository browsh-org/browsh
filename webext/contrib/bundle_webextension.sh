#!/bin/bash
# Convert the Browsh webextension into embedable binary data so that we can
# distribute Browsh as a single static binary.

# Requires the go-bindata binary, which seems to only be installed with:
#   `go get -u gopkg.in/shuLhan/go-bindata.v3/...`

set -e

PROJECT_ROOT=$(git rev-parse --show-toplevel)

NODE_BIN=$PROJECT_ROOT/webext/node_modules/.bin

cd $PROJECT_ROOT/webext && $NODE_BIN/webpack
cd $PROJECT_ROOT/webext/dist && $NODE_BIN/web-ext build --overwrite-dest

# Get the current version of Browsh
version=$(cat $PROJECT_ROOT/webext/manifest.json | python2 -c \
  'import sys, json; print json.load(sys.stdin)["version"]'
)

xpi_file=browsh-$version-an+fx.xpi
zip_file=browsh-$version.zip
source_dir=$PROJECT_ROOT/webext/dist/web-ext-artifacts

if [ "$BROWSH_ENV" == "RELEASE" ]
then
  # The signed version. There can only be one canonical XPI for each semantic
  # version.
  source_file=$source_dir/$xpi_file
  $webext sign --api-key $MDN_USER --api-secret $MDN_KEY
else
  # In development/testing, we want to be able to bundle the webextension
  # frequently without having to resort to version bumps.
  source_file=$source_dir/$zip_file
fi

bundle_file=$source_dir/browsh.xpi
destination=$PROJECT_ROOT/interfacer/webextension.go

cp -f $source_file $bundle_file

echo "Bundling $source_file to $destination..."

go-bindata \
  -nocompress \
  -prefix $PROJECT_ROOT \
  -pkg main \
  -o $destination \
  $bundle_file
