# How to setup Browsh's build system for Mac

If you want to try Browsh, you can use [Homebrew](https://brew.sh/).
Check out the [installation page](https://www.brow.sh/docs/installation/) at the
[official site](https://www.brow.sh/)).

## Installations

You need Go, Firefox and Node.js to run Browsh.

### Install Go

Follow the [installation guide](https://golang.org/doc/install) (you can use an installer).

### Install Firefox

Follow the official [guide](https://support.mozilla.org/en-US/kb/how-download-and-install-firefox-mac)
to install Firefox.

#### Include Firefox to your PATH

The `firefox` executable is probably at `/Applications/Firefox.app/Contents/MacOS`.
You need to add it to your `PATH` so that Browsh can create new instances of Firefox.

### Install Node.js

Follow the [official downloading page](https://nodejs.org/en/download/).

Use Nodejs > v8.11.4 with Browsh.

#### Install web-ext globally

It's a Mozilla tool for working with Firefox web extensions:

```shell
npm install -g web-ext
```

## Setting up your Browsh

### Clone Browsh

Fork Browsh to your Github account.
Clone it to a directory of your choice.
We will refer to this directory as `$browsh` for the rest of the guide.

### Install NPM packages

```shell
cd "$browsh/webext"
npm install
```

### Run the build script

```sh
cd "$browsh"
# install required package"
./interfacer/contrib/build_browsh.sh
```

## Running Browsh from source

Now that you have the required dependencies installed, we can run Browsh.
Open three terminals and do the following:

### Terminal 1 (builds JavaScript)

```sh
cd "$browsh/webext
# create a dist folder inside the webext folder.
npx webpack --watch
```

### Terminal 2 (handles Firefox web extension)

```sh
mkdir "$browsh/webext/dist"
cd "$browsh/webext/dist"
npx webpack --watch
```

### Terminal 3 (Displays Browsh)

```sh
cd "$browsh/interfacer"
go run ./cmd/browsh/main.go --firefox.use-existing --debug
```
