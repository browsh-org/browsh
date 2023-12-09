# How to setup the build environment for browsh on a generic Linux system

You can follow this guide as-is on x86-64 based Linux distributions.
You may try to adapt it to other systems.
This guide assumes that you cannot install the required Go, NodeJS and Firefox packages
from your distribution's repositories.
In case they are available, you should install them using your system's package manager.

## Installing golang

Get the latest binary for your target Linux system from the [golang download page](https://golang.org/dl/).

Extract to `/usr/local` with:

```shell
tar  -C /usr/local -xzf go1.11.linux-amd64.tar.gz
```

Add `/usr/local/go/bin` to your `PATH` in `~/.profile`

## Installing nodejs/npm

Go to the [nodejs download page](https://nodejs.org/download)
and select the LTS version of the Linux x86 64bit binaries.

```shell
mkdir /usr/local/lib/nodejs
VERSION=v8.11.4
DISTRO=linux-x64
tar -xJvf node-$VERSION-$DISTRO.tar.xz -C /usr/local/lib/nodejs
mv /usr/local/lib/nodejs/node-$VERSION-$DISTRO /usr/local/lib/nodejs/node-$VERSION
```

Edit your `~/.profile` to add `/usr/local/lib/nodejs/node-v8.11.4/bin` to your `PATH`,
then reload your profile:

```shell
source ~/.profile
```

Create symlinks for *node* and *npm*:

```shell
sudo ln -s /usr/local/lib/nodejs/node-$VERSION/bin/node /usr/local/bin/node
sudo ln -s /usr/local/lib/nodejs/node-$VERSION/bin/npm /usr/local/bin/npm
```

Now you should be able to call the *go* and *node* binaries:

```shell
go version
node --version
```

## Installing webpack and webpack-cli

The `--no-audit` flag gets around errors. You may not need it.

```shell
npm install -g --no-audit webpack
npm install -g webpack-cli
```

## Installing web-ext

The `--no-audit` flag gets around errors. You may not need it.

```shell
npm install -g --no-audit --ignore-scripts web-ext
```

## Installing Firefox

You may install *firefox* from your distribution's repositories.
**Browsh requires Version 57 or higher.**

### Installing firefox from mozilla's binaries

See `interfacer/contrib/setup_firefox.sh` for reference.

```shell
export FIREFOX_VERSION=60.0
mkdir -p $HOME/bin
pushd $HOME/bin
curl -L -o firefox.tar.bz2 https://ftp.mozilla.org/pub/firefox/releases/$FIREFOX_VERSION/linux-x86_64/en-US/firefox-$FIREFOX_VERSION.tar.bz2
bzip2 -d firefox.tar.bz2
tar xf firefox.tar
popd
```

## Cloning the browsh repository

It's assumed that you already have *git* installed.

Run this anywhere you want:

```shell
git clone https://github.com/browsh-org/browsh.git
```

## Setting up the build environment in the cloned repository

### Setting up dependencies

```shell
browsh=/path/to/browsh
cd "$browsh/webext"
source ~/.nvm/nvm.sh # this is optional
npm run get-gobindata
npm install
npm run build
diff -u <(echo -n) <(gofmt -d ./)
./node_modules/.bin/prettier --list-different "{src,test}/**/*.js"
```

### Building browsh

```shell
cd "$browsh/interfacer"
go build -o browsh src/main.go
```

### Building the web extension

In `$browsh/webext`:

```shell
webpack --watch
```

This will continuously watch for changes made to the web extension and rebuild it.

## Run firefox and the webextension

In `$browsh/webext/dist`:

```shell
web-ext run --verbose --firefox path/to/firefox
```

## Run browsh

```shell
cd "$browsh/interfacer"
go run ./cmd/browsh --firefox.use-existing --debug
```

Or after building:

```shell
./browsh --firefox.use-existing --debug
```
