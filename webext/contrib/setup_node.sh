#!/bin/bash

# See `/.nvmrc` for current Node version
NVM_VERSION=0.33.8
curl -o- https://raw.githubusercontent.com/creationix/nvm/v$NVM_VERSION/install.sh | bash
source $HOME/.nvm/nvm.sh
nvm install

