# How to setup the build environment for browsh on a generic Linux system

Install Go, Node.js, and Firefox using your system's package manager.
**Browsh requires Version 57 or higher.**

Now you should be able to call the *go* and *node* binaries:

```shell
go version
node --version
```

## Install webpack, webpack-cli, & web-ext

```shell
npm install -g webpack webpack-cli web-ext
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
