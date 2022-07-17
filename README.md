[![Follow @brow_sh](https://img.shields.io/twitter/follow/brow_sh.svg?style=social&label=Follow)](https://twitter.com/intent/follow?screen_name=brow_sh)

![Browsh Logo](https://www.brow.sh/assets/images/browsh-header.jpg)

**A fully interactive, real-time, and modern text-based browser rendered to TTYs and browsers**

![Browsh GIF](https://media.giphy.com/media/bbsmVkYjPdOKHhMXOO/giphy.gif)

## Why use Browsh?

Not all the world has good Internet.

If you only have a 3kbps internet connection tethered from a phone,
then it's good to SSH into a server and browse the web through, say,
[elinks](https://github.com/browsh-org/browsh/issues/17). That way the
_server_ downloads the web pages and uses the limited bandwidth of an
SSH connection to display the result. However, traditional text-based browsers
lack JS and all other modern HTML5 support. Browsh is different
in that it's backed by a real browser, namely headless Firefox,
to create a purely text-based version of web pages and web apps. These can be easily
rendered in a terminal or indeed, ironically, in another browser. Do note that currently the browser client doesn't have feature parity with the terminal client.

Why not VNC? Well VNC is certainly one solution but it doesn't quite
have the same ability to deal with extremely bad Internet. Terminal 
Browsh can also use MoSH to further reduce bandwidth and increase stability
of the connection. Mosh offers features like automatic
reconnection of dropped or roamed connections and diff-only screen updates.
Furthermore, other than SSH or MoSH, terminal Browsh doesn't require a client
like VNC.

One final reason to use terminal Browsh could be to offload the battery-drain of a modern
browser from your laptop or low-powered device like a Raspberry Pi. If you're a CLI-native,
then you could potentially get a few more hours of life if your CPU-hungry browser
is running somewhere else on mains electricity.

## Installation

Download a binary from the [releases](https://github.com/browsh-org/browsh/releases) (~7MB).
You will need to have [Firefox 63](https://www.mozilla.org/en-US/firefox/new/) (or higher) already installed.

Or download and run the Docker image (~230MB) with:
    `docker run --rm -it browsh/browsh`

## Usage
Most keys and mouse gestures should work as you'd expect on a desktop
browser.

For full documentation click [here](https://www.brow.sh/docs/introduction/).

## Contributing
_Note that some of these instructions may be out of date (July 2022)_

For generic Linux systems you can follow [this guide](https://github.com/browsh-org/browsh/blob/master/SETUP_DEV_LINUX.md) on how to setup a build environment, that you may be able to adapt for other systems as well.

Windows users can follow [this guide](https://github.com/browsh-org/browsh/blob/master/SETUP_DEV_WIN.md) in order to set up a build environment.

Mac users may follow [this guide](https://github.com/browsh-org/browsh/blob/master/SETUP_DEV_OSX.md) that goes through the steps of setting up a build environment.

### Communication
Questions about Brow.sh? Stuck trying to resolve a tricky issue? Try:
  - [Github Discussions](https://github.com/browsh-org/browsh/discussions)
  - [Gitter](https://gitter.im/browsh)
    
## Tests

For the webextension: in `webext/` folder, `npm test`    
For CLI unit tests: in `/interfacer` run `go test src/browsh/*.go`    
For CLI E2E tests: in `/interfacer` run `go test test/tty/*.go`    
For HTTP Service tests: in `/interfacer` run `go test test/http-server/*.go`    

## Special Thanks
  * [@tobimensch](https://github.com/tobimensch) For essential early feedback and user testing.
  * [@arasatasaygin](https://github.com/arasatasaygin) For the Browsh logo.

## Donating
Please consider donating: https://www.brow.sh/donate

## License
GNU Lesser General Public License v2.1
