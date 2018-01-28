#!/bin/bash

set -e

PROJECT_ROOT=$(git rev-parse --show-toplevel)
manifest=$PROJECT_ROOT/webext/manifest.json
line=$(cat $manifest | grep '"version"')
manifest_version=$(echo $line | grep -o '".*"' | cut -d " " -f 2 | sed 's/"//g')

latest_tagged_version=$(git tag --list 'v*.*.*' | head -n1 | sed -e "s/^v//")

if [[ "$manifest_version" != "$latest_tagged_version" ]]; then
  cd $PROJECT_ROOT/interfacer
  goreleaser
  git config --global user.email "builds@travis-ci.com"
  git config --global user.name "Travis CI"
  # `/dev/null` needed to prevent Github token appearing in logs
  git push --tags --quiet https://$GITHUB_TOKEN@github.com/tombh/texttop > /dev/null 2>&1
fi


