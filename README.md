# Browsh [![Build Status](https://travis-ci.org/tombh/texttop.svg?branch=master)](https://travis-ci.org/tombh/texttop)

**A fully interactive, realtime and modern browser rendered to TTY**

or Firefox in your terminal 😲

// Gifs go here

## Why?

Not all the world has good Internet.

If all you have is a 3kbps connection tethered from a phone
then it's good to SSH into a server and browse the web through
[elinks](https://github.com/tombh/texttop/issues/17). That way the
_server_ downloads the web pages and uses the limited bandwidth of an
SSH connection to display the result. But it lacks JS support and all
that other modern HTML5 goodness. So Browsh is simply a way to have
the power of a remote server running a modern browser, but interfaced
through the simplicity of a terminal and very low bandwidth.

Why not VNC? Well VNC is certainly one solution but it doesn't quite
have the same ability to deal with extremely bad Internet. Also,
Browsh can use MoSH to further reduce the bandwidth and stability
requirements of the connection. Mosh offers features like automatic
reconnection of dropped connections and diff-only screen updates.
Furthermore, other than SSH or MoSH, Browsh doesn't require a client
like VNC.

Another reason could be to offload the battery-drain of a modern
browser from your laptop. If you're a CLI-native, then you could
potentially get a few more hours life if your CPU-hungry browser
is running somehwere else on mains electricity.

But of course the biggest reason for Browsh is probably just that it's
cool geekery. You may just appreciate the sheer simplicty of browsing
a text-based web in your terminal.

## Installation

Download a [https://github.com/tombh/browsh/releases](release) (~2MB).
You will need to have Firefox >=57 aleady installed.

Or download and run the Docker image (~800MB) with:
    `docker run -it tombh/browsh`

There's currently a known bug when running Browsh against a fresh install of Firefox.
If Browsh happens to be the first to ever open Firefox on your machine then it will
get stuck because Firefox behaves differently on its first ever launch. The simple
fix is to either just to run Browsh again, or launch Firefox manually yourself (you don't even need to keep it open).

## Usage
Most keys and mouse gestures should work as you'd expect on a desktop
browser.

For full documentation see [the documentation](https://www.brow.sh/docs/introduction/)

## Contributing
To setup a dev env you will need NodeJS and Golang installed. If you get stuck
setting up your env, take a look in `.travis.yml`, it has to setup everything
from scratch for every push to Github.

I'd recommend `[nvm](https://github.com/creationix/nvm)` for NodeJS - note that
`nvm install` will automatically parse the `.nvmrc` version in this repo to get
the correct NodeJS version. For Golang it's probably best to just use your OS's
package manager. The current Golang version being used is stored in `.travis.yml`.

You'll then need to install the project dependencies. For the webextension, just
run: `npm install` inside the `webext/` folder. For the CLI client you will first
need to install `dep`, there is a script for this in `interfacer/contrib/setup_go.sh`.
I don't fully understand Golang's best practices, but it seems you are forced to
keep your Go project's code under `$GOPATH/src`, you might be able to get away
with symlinks. Anyway, to install the dependencies use: `dep ensure` inside the
`interfacer/` folder.

Then the ideal setup for development is:
  * have Webpack watch the JS code so that it rebuilds automatically:
    `webpack --watch`
  * run the CLI client without giving it the responsibility to launch Firefox:
    `go run ./interfacer/*.go -use-existing-ff`
  * have Mozilla's handy `web-ext` tool run Firefox and reinstall the
    webextension everytime webpack rebuilds it: (in `webext/dist`)
    `web-ext run --verbose --url https://google.com`

## Tests

For the webextension: in the `webext/` folder, `npm test`
For the CLI client: in the `interfacer/` folder `./contrib/run_tests.sh`

## License
GNU General Public License v3.0
