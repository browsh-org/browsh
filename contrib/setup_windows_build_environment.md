# How to set up the build environment for Browsh on Windows


## Installing golang
Get the latest amd64 for Windows on the [golang download page](https://golang.org/dl/).

Run the msi file.

## Installing nodejs/npm

Go to the [nodejs download page](https://nodejs.org/en/download/) and select the LTS version of Windows Installer (msi).

Run the msi file.

## Setting up GOPATH

Create a go directory:
> mkdir go

> cd go

Set GOPATH to current directory.
> set GOPATH=%cd%

Create subdirectories bin and src within your go directory:
> mkdir bin
> mkdir src


## Installing dep

Using chocolatey package manager run:
> choco install dep


## Installing webpack
> npm install -g --no-audit webpack

## Installing web-ext
npm install -g --ignore-scripts web-ext

## Installing firefox
**Version 57 or higher is required.**


## Cloning the browsh repository
Navigate to GOPATH/src and run:
git clone https://github.com/browsh-org/browsh.git

## Setting up the build environment in the cloned repository

### Setting up dependencies

Navigate to interfacer and run:
> dep ensure

Navigate to webext and run:
> npm run get-gobindata


