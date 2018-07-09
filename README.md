[![Build Status](https://travis-ci.org/browsh-org/browsh.svg?branch=master)](https://travis-ci.org/browsh-org/browsh)

![Browsh Logo](https://www.brow.sh/assets/images/browsh-header.jpg)

**A fully interactive, realtime and modern text-based browser rendered to TTYs and browsers**

![Browsh Gif](https://media.giphy.com/media/bbsmVkYjPdOKHhMXOO/giphy.gif)

## Why?

Not all the world has good Internet.

If all you have is a 3kbps connection tethered from a phone
then it's good to SSH into a server and browse the web through, say,
[elinks](https://github.com/browsh-org/browsh/issues/17). That way the
_server_ downloads the web pages and uses the limited bandwidth of an
SSH connection to display the result. But traditional text-based browsers
lack JS support and all that other modern HTML5 goodness. Browsh is different
in that it's backed by a real browser, namely headless Firefox, and uses that
to create purely text-based version of web pages and web apps that can be easily
rendered in a terminal or indeed, somewhat ironically, in another browser. Though note that currently the browser client doesn't have feature parity with the terminal client.

Why not VNC? Well VNC is certainly one solution but it doesn't quite
have the same ability to deal with extremely bad Internet. Also,
terminal Browsh can use MoSH to further reduce bandwidth and increase stability
of the connection. Mosh offers features like automatic
reconnection of dropped or roamed connections and diff-only screen updates.
Furthermore, other than SSH or MoSH, terminal Browsh doesn't require a client
like VNC.

One final reason to use terminal Browsh could be to offload the battery-drain of a modern
browser from your laptop or low-powered device like a Raspberry Pi. If you're a CLI-native,
then you could potentially get a few more hours life if your CPU-hungry browser
is running somewhere else on mains electricity.

## Installation

Download a binary from the [releases](https://github.com/browsh-org/browsh/releases) (~2-6MB).
You will need to have Firefox >=57 aleady installed.

Or download and run the Docker image (~230MB) with:
    `docker run -it browsh/browsh`

## Usage
Most keys and mouse gestures should work as you'd expect on a desktop
browser.

For full documentation see [the documentation](https://www.brow.sh/docs/introduction/).

## Contributing
To setup a dev env you will need NodeJS and Golang installed. If you get stuck
setting up your env, take a look in `.travis.yml`, it has to setup everything
from scratch for every push to Github.

I'd recommend [nvm](https://github.com/creationix/nvm) for NodeJS - note that
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
    `go run ./interfacer/src/main.go -use-existing-ff -debug`
  * have Mozilla's handy `web-ext` tool run Firefox and reinstall the
    webextension everytime webpack rebuilds it: (in `webext/dist`)
    `web-ext run --verbose`

## Tests

For the webextension: in `webext/` folder, `npm test`    
For CLI unit tests: in `/interfacer` run `go test src/browsh/*.go`    
For CLI E2E tests: in `/interfacer` run `go test test/tty/*.go`    
For HTTP Service tests: in `/interfacer` run `go test test/http-server/*.go`    

## Special Thanks
  * [@tobimensch](https://github.com/tobimensch) For essential early feedback and user testing.
  * [@arasatasaygin](https://github.com/arasatasaygin) For the Browsh logo.

## License
GNU General Public License v3.0
