# Browsh
**A fully interactive, realtime and modern browser rendered to TTY**

or Firefox in your terminal 😲

## Why?
I'm travelling around the world and sometimes I don't have very good
Internet. If all I have is a 3kbps connection tethered from my phone
then it's good to SSH into my server and browse the web through
[elinks](http://www.xteddy.org/elinks/). That way my _server_ downloads
the web pages and uses the limited bandwidth of my SSH connection to
display the result. But it lacks JS support and all that other modern
HTML5 goodness. Browsh is simply a way to have the power of a remote
server running a desktop, but interfaced through the simplicity of a
terminal and very low bandwidth.

Why not VNC? Well VNC is certainly one solution but it doesn't quite
have the same ability to deal with extremely bad Internet. Browsh can
use MoSH to further reduce the bandwidth and stability requirements
of the connection. Mosh offers features like automatic reconnection
of dropped connections and diff-only screen updates. Also, other than
SSH or MoSH, Browsh doesn't require a client like VNC. But of course
another big reason for Browsh is that it's just very cool geekery.

## Installation

Download a [](release). You will need to have Firefox 57+ aleady
installed.

Or download and run the Docker image with
    `docker run -it tombh/browsh`

## Usage
Most keys and mouse gestures should work as you'd expect on a desktop
browser.

`CTRL+l` Focus URL bar

## License
GNU General Public License v3.0
