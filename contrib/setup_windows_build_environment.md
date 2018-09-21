# How to set up the build environment for Browsh on Windows
This guide is for those who want to set up the build environment on Windows Command Prompt or Powershell. Since some of the shell scripts are needed to set up the environment, you can just use **Git Bash** to run these scripts.


## Setting up Go, NodeJs, and GOPATH
Download and install Go at [Go download page](https://golang.org/dl/).

Download and install NodeJs at [NodeJs download page](https://nodejs.org/en/download/)

Using Command Prompt or Powershell:

Create a go directory:
> mkdir go

> cd go

Set GOPATH to current directory.
> set GOPATH=%cd%

Create subdirectories bin and src within your go directory:
> mkdir bin

> mkdir src

## Installing chocolatey and dep

Download and install Chocolatey package manager at [Chocolatey download page](https://chocolatey.org/install).

Using chocolatey package manager run:
> choco install dep


## Installing webpack, web-ext, and Firefox
> npm install -g --no-audit webpack

> npm install -g --ignore-scripts web-ext

Download and install Firefox for Windows at [Firefox download page](https://www.mozilla.org/en-US/firefox/new/).
Note: **Version 57 or higher is required.**


## Cloning the browsh repository
Navigate to GOPATH/src and run:
git clone https://github.com/browsh-org/browsh.git


## Setting up dependencies

Navigate to browsh/webext and run:
> npm run get-gobindata

> npm install

## Building Browsh

Using Git Bash, navigate to browsh/interfacer/contrib and run:

> ./build_browsh.sh

> ./xpi2bin.sh

## Running Browsh

Using Command Prompt or Powershell:

Navigate to GOPATH/browsh and run:
go run ./interfacer/src/main.go --firefox.use-existing --debug

Navigate to browsh/webext and run:
webpack --watch

Navigate to browsh/webext/dist and run:
web-ext run --verbose





