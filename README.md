#Texttop
**A fully interactive X Linux desktop rendered to ASCII and streamed over SSH**

or Firefox in your terminal 😲

[gif here]

##Why?
I'm travelling around the world and sometimes I don't have very good Internet. If all I have is a 3kbps connection
tethered from my phone then it's good to SSH into my server and browse the web through [elinks](http://www.xteddy.org/elinks/).
That way my server downloads the web pages and uses the limited bandwidth of my SSH connection to display the result. But
it lacks JS support and all that other modern HTML5 goodness. Texttop is simply a way to have the power of a remote
server running a desktop, but interfaced through the simplicity of a terminal and very low bandwidth.

Why not VNC? Well VNC is certainly one solution but it doesn't quite have the same ability to deal with extremely bad
Internet. Texttop uses MoSH to further reduce the bandwidth and stability of the connection. Mosh offers features like
automatic reconnection of dropped connections and diff-only screen updates. Also, other than SSH or MoSH, Texttop doesn't
require a client like VNC. But of course another big reason for Texttop is that it's just very cool geekery.

##Quickstart
If you just want to have a play on your local machine:
```
docker run --rm -it tombh/texttop sh
./run.sh
```

##Installation
You can either pull from the Docker Registry:
`docker pull tombh/textop`
or, build yourself:
```
git clone git@github.com:tombh/texttop.git
cd texttop
docker build -t texttop .
```

##Usage
On your remote server (this will pull the docker image the first time you issue it):
```
docker run --rm -it \
  -p 7777:7777 -p 60010-60020:60010-60020/udp \
  -v ~/.ssh/authorized_keys:/root/.ssh/authorized_keys \
  tombh/texttop sh
```
Note that this assumes you already have SSH setup on your server and that you have your public key there. Password
logins work fine too.

Then on your local machine:
```
mosh user@yourserver:7777
./run.sh
```
MoSH is available through most system pacakge managers. SSH can be used exactly the same, just replace `mosh` with `ssh`.
`user@yourserver` is the normal URI you would use to connect via SSH.

**Exiting**
At the moment the only way to exit is with MoSH's `CTRL+^ .` or SSH's `ENTER ~ .`

##Interaction
Most mouse and keyboard input is exactly the same as a normal desktop. If your terminal is active then you can click,
type, scroll, use arrow keys and drag things around. However there are still some things not available, like copy and
paste. The main difference from a normal desktop is that you can zoom and pan the desktop by using `CTRL+mousewheel` and
`CTRL+drag`. This is very handy as it's hard to see what's what when you're zoomed right out.

Currently, only Firefox is installed on this extremely minimal Alpine Linux distro. However you can add new packages
with [apk](https://wiki.alpinelinux.org/wiki/Alpine_Linux_package_management). Just remember that you will lose any
system changes once you restart the docker container. I'm thinking about ways to save state. You may experiment with
mounting certain system directories.

##Contributions
Yes please.

##License
GNU General Public License v3.0
