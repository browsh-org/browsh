# How to setup Browsh's build system for Mac
If you just want to try Browsh, you can use [Homebrew](https://brew.sh/) (check out the [installation page](https://www.brow.sh/docs/installation/) at the [official site](https://www.brow.sh/)).

## Installations
You need Go, Firefox and Node.js to run Browsh.

### Install Go
Follow the [installation guide](https://golang.org/doc/install) (you can use an installer).

#### Ensure your GOPATH is set

```sh
$ echo $GOPATH
/Users/uesr_name/go
$ # anywhere is ok, but make sure it's not none
```

#### Ensure you have `$GOPATH/src` and `$GOPATH/bin` folders
If you're not sure if you have these folders, run:

```sh
$ mkdir "$GOPATH/src"
$ mkdir "$GOPATH/bin"
```

### Install Firefox
Follow the official [guide](https://support.mozilla.org/en-US/kb/how-download-and-install-firefox-mac) to install Firefox.

#### Include Firefox to your PATH
The `firefox` executable is probably at `/Applications/Firefox.app/Contents/MacOS`. You need to add it to your `PATH` so that Browsh can create new instances of Firefox.

### Install Node.js
Follow the [official downloading page](https://nodejs.org/en/download/).

> v8.11.4. is currently recommended for working with Browsh (?)

#### Install web-ext globally
It's a Mozilla's handy tool for working with Firefox web extensions:

```sh
$ npm install -g web-ext
```

## Setting up your Browsh

### Clone Browsh
Fork Browsh to your Github account. Clone it to `$GOPATH/src`.

### Install NPM packages

```shell
$ cd "$GOPATH/src/browsh/webext"
$ npm install
```

### Run the build script

```sh
$ cd "$GOPATH/src/browsh"
$ # install several required package"
$ ./interfacer/contrib/build_browsh.sh
```

## Running Browsh from source
Now that you have all of the required dependencies installed, we can run Browsh. Open three terminals and do the follows:

### Terminal 1 (builds JavaScript)

```sh
$ cd "$GOPATH/src/browsh/webext"
$ # create a dist folder inside the webext folder.
$ npx webpack --watch
```

### Terminal 2 (handles Firefox web extension)

```sh
$ # the dist folder is created in the first terminal
$ cd "$GOPATH/browsh/webext/dist"
$ # create a dist folder inside the webext folder.
$ npx webpack --watch
```

### Terminal 3 (Displays Browsh)

```sh
$ cd "$GOPATH/browsh"
$ go run ./interfacer/src/main.go --firefox.use-existing --debug
```

