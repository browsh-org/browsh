# How to set up the build environment for Browsh on Windows

This guide is for those who want to set up the build environment on Windows
Command Prompt or Powershell.
Setup depends on running shell scripts. You can use **Git Bash** to run those scripts.

## Setting up Go, NodeJs, and GOPATH

Download and install Go for Windows at [Go download page](https://golang.org/dl/).

Download and install NodeJs for Windows at [NodeJs download page](https://nodejs.org/en/download/)

Using Command Prompt or Powershell:

Create a go workspace:

```shell
mkdir go
cd go
```

Set GOPATH to current directory.

```shell
set GOPATH=%cd%
```

Create subdirectories bin and src within your go directory:

```shell
mkdir bin
mkdir src
```

Add %GOPATH%/bin to your PATH.

## Installing chocolatey and dep

Download and install Chocolatey package manager at [Chocolatey download page](https://chocolatey.org/install).

Using chocolatey package manager run:

```shell
choco install dep
```

## Installing webpack, web-ext, and Firefox

```shell
npm install -g --no-audit webpack
npm install -g --ignore-scripts web-ext
```

Download and install Firefox for Windows at [Firefox download page](https://www.mozilla.org/en-US/firefox/new/).
Note: **Browsh requires Firefox versions 57 or higher.**

## Cloning the browsh repository

Navigate to GOPATH/src and run:

```shell
git clone https://github.com/browsh-org/browsh.git
```

## Setting up dependencies

Navigate to browsh/webext and run:

```shell
npm install
```

## Building browsh with Git Bash

Using Git Bash, navigate to browsh/interfacer/contrib and run:

```shell
./build_browsh.sh
```

## Running browsh

Using three Command Prompts or Powershells:

Navigate to GOPATH/browsh and run:

```shell
go run ./interfacer/src/main.go --firefox.use-existing --debug
```

Navigate to browsh/webext and run:

```shell
webpack --watch
```

Navigate to browsh/webext/dist and run:

```shell
web-ext run --verbose
```
