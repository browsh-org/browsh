#!/bin/bash
# Convert the Browsh webextension into embedable binary data so that we can
# distribute Browsh as a single static binary.

# Requires the go-bindata binary, which seems to only be installed with:
#   `go get -u gopkg.in/shuLhan/go-bindata.v3/...`

set -e

PROJECT_ROOT=$(git rev-parse --show-toplevel)

NODE_BIN=$PROJECT_ROOT/webext/node_modules/.bin
destination=$PROJECT_ROOT/interfacer/src/browsh/webextension.go

cd $PROJECT_ROOT/webext && $NODE_BIN/webpack
cd $PROJECT_ROOT/webext/dist && rm *.map
ls -alh .
$NODE_BIN/web-ext build --overwrite-dest
ls -alh web-ext-artifacts

version=$($PROJECT_ROOT/contrib/get_browsh_version.sh)

xpi_file=browsh-$version-an+fx.xpi
zip_file=browsh-$version.zip
source_dir=$PROJECT_ROOT/webext/dist/web-ext-artifacts

if [ "$BROWSH_ENV" == "RELEASE" ]
then
  # The signed version. There can only be one canonical XPI for each semantic
  # version.
  source_file=$source_dir/$xpi_file
  bundle_file=$source_dir/browsh.xpi
  $NODE_BIN/web-ext sign --api-key $MDN_USER --api-secret $MDN_KEY
else
  # TODO: This doesn't currently work with the Marionettte `tempAddon`
  # installation method. Just use `web-ext run` and Browsh's `use-existing-ff`
  # flag - which is better anyway as it auto-reloads the extension when files
  # change. NB: If you fix this, don't forget to change the filename loaded
  # by `Asset()` in `main.go`.
  # In development/testing, we want to be able to bundle the webextension
  # frequently without having to resort to version bumps.
  source_file=$source_dir/$zip_file
  bundle_file=$source_dir/browsh.zip
fi

cp -f $source_file $bundle_file

echo "Bundling $source_file to $destination using internal path $bundle_file"

go-bindata -version
go-bindata \
  -nocompress \
  -prefix $PROJECT_ROOT \
  -pkg browsh \
  -o $destination \
  $bundle_file

ls -alh $PROJECT_ROOT/interfacer/src/browsh/webextension.go
echo "go-bindata exited with $(echo $?)"
