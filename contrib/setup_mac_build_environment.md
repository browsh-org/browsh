# How to setup Browsh's build system for Mac

## Install Go
Follow the [install guide](https://golang.org/doc/install). Note that there is an installer for Mac.

### Ensure your GOPATH is set
Open your terminal of choice. Run `echo $GOPATH`.
You should see something like `/usr/local/go`. Inside this folder, there will be a src folder. If there isn't one created yet, run `mkdir $GOPATH/src`.

## Clone Browsh
Fork Browsh to your Github account. Clone this fork of Browsh to your $GOPATH/src folder you just created.

## Install Firefox
Follow Firefox's [guide](https://support.mozilla.org/en-US/kb/how-download-and-install-firefox-mac) to installing Firefox on Mac.

### Add the Firefox app to your PATH
Browsh needs to be able to create new instances of Firefox. Add the Firefox app to your path. It's probably something like `/Applications/Firefox.app/Contents/MacOS`.
To add this to your path, edit your path file by running `sudo nano /etc/paths`. Add the path to Firefox in here and save the file.

## Install Node
[Install Node](https://nodejs.org/en/download/). The currently recommended version of Node for working with Browsh is v8.11.4.

### Install NPM packages
Navigate to browsh/webext. Run `npm install`.

### Install web-ext globally
Run `npm install -g web-ext`. This is Mozilla's handy tool for working with Firefox web extensions.

## Run the build script
Navigate to the root of your Browsh project. This should be `$GOROOT/src/browsh`. Run `./interfacer/contrib/build_browsh.sh`. This will install several required packages.

## Running Browsh from source
Now that you have all of the required dependencies installed, we can run Browsh. First, open 3 terminals.

### Terminal 1
This terminal will build the Javascript. From the `browsh/webext` folder, run `npx webpack --watch`. This will create a dist folder inside the webext folder.

### Terminal 2
This terminal will handle the Firefox web extension. From the `browsh/webext/dist` folder, run `web-ext run --verbose`.

### Terminal 3
This terminal will display Browsh. From the project root, run `go run ./interfacer/src/main.go --firefox.use-existing --debug`. 
