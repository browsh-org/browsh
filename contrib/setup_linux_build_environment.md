# How to setup the build environment for browsh on a generic linux system

This guide was made for x86-64 based Linux systems. You may try to adapt it to other systems.
In this guide it is assumed that you can't install the required go, nodejs and firefox versions from your distribution's repositories.In case this shouldn't be correct, you should try to install go, nodejs and firefox using the package manager of your distribution.

## Installing golang

Get the latest amd64 binary for Linux on the [golang download page](https://golang.org/dl/).

Extract to `/usr/local` with:
> tar  -C /usr/local -xzf go1.11.linux-amd64.tar.gz

Add `/usr/local/go/bin` to your `PATH` in `~/.profile`

## Installing nodejs/npm

Go to the [nodejs download page](https://nodejs.org/download) and select the LTS version of the Linux x86 64bit binaries.

> mkdir /usr/local/lib/nodejs

> VERSION=v8.11.4

> DISTRO=linux-x64

> tar -xJvf node-$VERSION-$DISTRO.tar.xz -C /usr/local/lib/nodejs

> mv /usr/local/lib/nodejs/node-$VERSION-$DISTRO /usr/local/lib/nodejs/node-$VERSION

Edit your `~/.profile` and add `/usr/local/lib/nodejs/node-v8.11.4/bin` to your `PATH`.

Reload your profile:
> source ~/.profile

Create symlinks for *node* and *npm*:
> sudo ln -s /usr/local/lib/nodejs/node-$VERSION/bin/node /usr/bin/node

> sudo ln -s /usr/local/lib/nodejs/node-$VERSION/bin/npm /usr/bin/npm

Now you should be able to call the *go* and *node* binaries:
> go --version

> node --version

## Setting up GOPATH
Add these two lines to the end of your *bash* profile in ~/.profile:
> GOPATH=$HOME/go/

> PATH="$PATH:$GOPATH/bin"

Reload your profile:
> source ~/.profile

Make sure the path `GOPATH/bin` exists with:
> mkdir -p $GOPATH/bin

## Installing dep (dependency manager for go)
See script in `browsh/interfacer/contrib/setup_dep.sh` for reference.

Get the latest amd64 linux binary of *dep* at [https://github.com/golang/dep/releases/]()

Move the binary to `GOPATH/bin/dep` and make sure `GOPATH/bin` exists with `mkdir -p $GOPATH/bin`
> chmod +x $GOPATH/bin/dep

## Installing webpack
(--no-audit is used to get around errors, may not be needed)
> npm install -g --no-audit webpack

## Installing web-ext
(--ignore-scripts is used to get around errors, may not be needed)
> npm install -g --ignore-scripts web-ext

## Installing firefox
You may install *firefox* from your distribution's repositories. **Version 57 or higher is required.**

### Installing firefox from mozilla's binaries
See `interfacer/contrib/setup_firefox.sh` for reference.

> export FIREFOX_VERSION=60.0

> mkdir -p $HOME/bin

> pushd $HOME/bin

> curl -L -o firefox.tar.bz2 https://ftp.mozilla.org/pub/firefox/releases/$FIREFOX_VERSION/linux-x86_64/en-US/firefox-$FIREFOX_VERSION.tar.bz2

> bzip2 -d firefox.tar.bz2

> tar xf firefox.tar

> popd

## Cloning the browsh repository
It's assumed that you already have *git* installed.

Ensure `GOPATH/src` exists with:
> mkdir -p "$GOPATH/src"

Run this in `GOPATH/src`:
> git clone https://github.com/browsh-org/browsh.git

## Setting up the build environment in the cloned repository

### Setting up dependencies 

> REPO_ROOT=$GOPATH/src/browsh

> cd $REPO_ROOT/interfacer && dep ensure

> cd $REPO_ROOT/webext

(source ~/.nvm/nvm.sh)

> npm run get-gobindata

> npm install

> npm run build

> diff -u <(echo -n) <(gofmt -d ./)

> ./node_modules/.bin/prettier --list-different "{src,test}/**/*.js"

### Building browsh

> cd $REPO_ROOT/interfacer

> go build -o browsh src/main.go

### Building the web extension

In `REPO_ROOT/webext`:
> webpack --watch

This will continuously watch changes made to the web extension and rebuild it.

## Run firefox and the webextension
In `REPO_ROOT/webext/dist`:
> web-ext run --verbose --firefox path/to/firefox

## Run browsh
> go run ./interfacer/src/main.go --firefox.use-existing --debug

Or after building:
> ./browsh --firefox.use-existing --debug

