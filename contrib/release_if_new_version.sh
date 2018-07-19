#!/bin/bash

set -e

PROJECT_ROOT=$(git rev-parse --show-toplevel)

browsh_version=$($PROJECT_ROOT/contrib/get_browsh_version.sh)
latest_tagged_version=$(git tag --sort=v:refname --list 'v*.*.*' | tail -n1 | sed -e "s/^v//")

echo "Browsh version: $browsh_version"
echo "Latest tag: $latest_tagged_version"

if [[ "$browsh_version" == "$latest_tagged_version" ]]; then
  echo "Not running release as there's no new version."
  exit 0
fi

git tag v$browsh_version
git show v$browsh_version --quiet
git config --global user.email "builds@travis-ci.com"
git config --global user.name "Travis CI"
# `/dev/null` needed to prevent Github token appearing in logs
git push --tags --quiet https://$GITHUB_TOKEN@github.com/browsh-org/browsh > /dev/null 2>&1

git reset --hard v$browsh_version

cd $PROJECT_ROOT/webext
BROWSH_ENV=RELEASE npm run build

cd $PROJECT_ROOT/interfacer/src
curl -sL http://git.io/goreleaser | bash

cd $HOME
git clone https://github.com/browsh-org/www.brow.sh.git
cd www.brow.sh
echo "latest_version: $browsh_version" > _data/browsh.yml
git add _data/browsh.yml
git commit -m "(Travis CI) Updated Browsh version to $browsh_version"
# `/dev/null` needed to prevent Github token appearing in logs
git push --quiet https://$GITHUB_TOKEN@github.com/browsh-org/www.brow.sh > /dev/null 2>&1
