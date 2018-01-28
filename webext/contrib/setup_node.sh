#!/bin/bash

if ! type "nvm" > /dev/null; then
  rm -rf ~/.nvm
  NVM_VERSION=0.33.8
  curl -o- https://raw.githubusercontent.com/creationix/nvm/v$NVM_VERSION/install.sh | bash
  source $HOME/.nvm/nvm.sh
fi

nvm install # See `/.nvmrc` for current Node version

