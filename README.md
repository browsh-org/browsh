# Browsh [![Build Status](https://travis-ci.org/tombh/texttop.svg?branch=master)](https://travis-ci.org/tombh/texttop)

**A fully interactive, realtime and modern browser rendered to TTY**

or Firefox in your terminal 😲

// Gifs go here

## Why?

I'm travelling around the world and sometimes I don't have very good
Internet. If all I have is a 3kbps connection tethered from my phone
then it's good to SSH into my server and browse the web through
[elinks](https://github.com/tombh/texttop/issues/17). That way my
_server_ downloads the web pages and uses the limited bandwidth of my
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
You will need to have Firefox 57+ aleady installed.

Or download and run the Docker image (~800MB) with:
    `docker run -it tombh/browsh`

## Usage
Most keys and mouse gestures should work as you'd expect on a desktop
browser.

`CTRL+l` Focus URL bar

## License
GNU General Public License v3.0
