Hello newcomers. This has suddenly got some unplanned exposure. I'm actually in the middle of a complete rewrite on the [webext-rewrite branch](https://github.com/tombh/texttop/tree/webext-rewrite). You can get an idea of my approach with this function: https://github.com/tombh/texttop/blob/webext-rewrite/webext/src/text_builder.js#L165

After the success of hitting the front page of Hacker News last year, I really wanted to sit down and turn this proof of concept into something solid. So not only am I working on real text support (that you can of course copy and paste without even zooming). But I've removed the dependencies on `ffmpeg`, Xorg (for Firefox at least - Chrome strangely doesn't support webextensions in headless mode), `docker` AND it will work on all webextension-compatible browsers. It's going to be a single cross-platform, static Go binary, that launches your preferred browser in the background.

Generally it's bad luck to talk about something before it's finished, but seeing as it's suddenly getting attention again I wanted to let all those interested know that I'm making this 10 times better.

# Texttop
**A fully interactive X Linux desktop rendered to TTY and streamed over SSH**

or Firefox in your terminal 😲

![Alt Text](https://i.imgur.com/jX3vhO4.gif)

This [Youtube video](https://www.youtube.com/watch?v=TE_D_fx_ut8) gives a more faithful rendition of the experience.

## Why?
I'm travelling around the world and sometimes I don't have very good Internet. If all I have is a 3kbps connection
tethered from my phone then it's good to SSH into my server and browse the web through [elinks](https://github.com/tombh/texttop/issues/17).
That way my _server_ downloads the web pages and uses the limited bandwidth of my SSH connection to display the result. But
it lacks JS support and all that other modern HTML5 goodness. Texttop is simply a way to have the power of a remote
server running a desktop, but interfaced through the simplicity of a terminal and very low bandwidth.

Why not VNC? Well VNC is certainly one solution but it doesn't quite have the same ability to deal with extremely bad
Internet. Texttop uses MoSH to further reduce the bandwidth and stability requirements of the connection. Mosh offers features like
automatic reconnection of dropped connections and diff-only screen updates. Also, other than SSH or MoSH, Texttop doesn't
require a client like VNC. But of course another big reason for Texttop is that it's just very cool geekery.

## Quickstart
If you just want to have a play on your local machine:
```
docker run --rm -it tombh/texttop sh
./run.sh
```

## Installation
You can either pull from the Docker Registry:
`docker pull tombh/texttop`
or, build yourself:
```
git clone https://github.com/tombh/texttop.git
cd texttop
docker build -t texttop .
```
The docker image is only ~275MB.

## Usage
On your remote server (this will pull the docker image the first time you issue it):
```
docker run -d \
  -p 7777:7777 -p 60000-60020:60000-60020/udp \
  -v ${HOME}/.ssh:/root/.ssh:ro \
  tombh/texttop
```
Note that this assumes you already have SSH setup on your server and that you have your public key there. Password
logins work fine too. The `60000-60020` port range is for MoSH.

Then on your local machine:
```
mosh --ssh="ssh -p 7777" user@yourserver
cd /app
./run.sh
```
MoSH is available through most system package managers. SSH can be used exactly the same, just replace `mosh` with `ssh`.

`user@yourserver` is the normal URI you would use to connect via SSH.

**Exiting**    
`CTRL+ALT+Q` will drop you back to the docker container's CLI. You can start again with `./run.sh`

If MoSH or SSH become unresponsive you can exit MoSH with `CTRL+^ .` or SSH with `ENTER ~ .`

## Interaction
  * `CTRL + mousewheel` to zoom
  * `CTRL + click/drag` to pan

Most mouse and keyboard input is exactly the same as a normal desktop. If your terminal is active then you can click,
type, scroll, use arrow keys and drag things around. However there are still some things not available, like copy and
paste. The main difference from a normal desktop is that you can zoom and pan the desktop by using `CTRL + mousewheel` and
`CTRL + drag`. This is very handy as it's hard to see what's what when you're zoomed right out.

### Keyboard Mode
If your terminal doesn't support mouse input then you can switch in and out of keyboard mode with `CTRL+ALT+M`.
This will give you the following shortcuts:

`u` mouse up    
`n` mouse down    
`h` mouse left    
`k` mouse right    

`SHIFT+u` pan up    
`SHIFT+n` pan down    
`SHIFT+h` pan left    
`SHIFT+k` pan right    

`CTRL+u` zoom in    
`CTRL+n` zoom out    

`j` left-click    
`r` right-click    
`t` middle-click    

### Adding new applications
Currently, only Firefox is installed on this extremely minimal Alpine Linux distro. However you can add new packages
with [apk](https://wiki.alpinelinux.org/wiki/Alpine_Linux_package_management). Example;
```
# Login with a separate session
apk --no-cache add xterm
export DISPLAY=:0
xterm &
```
Just remember that you will lose any system changes once you restart the docker container. I'm thinking about ways to
save state. You may experiment with mounting certain system directories. Eg;
```
docker run -d \
  -p 7777:7777 -p 60000-60020:60000-60020/udp \
  -v ${HOME}/.ssh:/root/.ssh:ro \
  -v ${HOME}/.texttop/var:/var \
  tombh/texttop
```

## Known Issues
The Docker Hub version is built against Intel CPU architectures, this causes hiptext to fail on AMD chips. In which
case you will need to build texttop yourself:
```
git clone https://github.com/tombh/texttop.git
cd texttop
docker build -t texttop .
```

**Working terminals**
  * [Tilda](https://github.com/lanoxx/tilda)
  * [Terminal](https://launchpad.net/pantheon-terminal)

**Problematic terminals**
  * konsole: neither `CTRL+click/drag` nor `CTRL+mousewheel` are forwarded (perhaps mouse reporting is disabled by default)
  * xterm: `CTRL+click/drag` is intercepted by the GUI menu
  * rxvt: rendering issues

## Contributions
Yes please.

## License
GNU General Public License v3.0
