#!/bin/sh

export LC_ALL=C
export LANG=C

export DESKTOP_WIDTH='1600'
export DESKTOP_HEIGHT='1200'
export DISPLAY=:0
DESKTOP_RES="$DESKTOP_WIDTH"x"$DESKTOP_HEIGHT"
UDP_URI='udp://127.0.0.1:1234'

# Create an X desktop in memory without actually displaying it on a real screen
Xvfb :0 -screen 0 "$DESKTOP_RES"x16 > xvfb.log 2>&1 &

# TODO: detect X start rather than sleep
sleep 1

/usr/bin/startxfce4 >> xvfb.log 2>&1 &
/usr/bin/thunar >> xvfb.log 2>&1 &

# Convert the X framebuffer desktop into a video stream
ffmpeg \
  -f x11grab \
  -s $DESKTOP_RES \
  -r 12 \
  -i :0.0 \
  -vcodec mpeg2video \
  -f mpegts \
  $UDP_URI \
  > ffmpeg.log 2>&1 &

# The above ffmpeg can take a while to open the UDP stream, so wait a little
# TODO: detect the stream's presence rather than sleep
sleep 1

# Intercept STDIN (mouse and keypresses) and forward to the X framebuffer via xdotool
(./stdin_forward <&3) 3<&0

# Hiptext renders images and videos into text characters displayable in a terminal
# Hiptext complains unless you specify the exact path to the font, seems like a bug to me
# TODO: support variable width, ideally dynamic sizing
# hiptext \
#   -font /usr/share/fonts/ttf-dejavu/DejaVuSansMono.ttf \
#   $UDP_URI \
#   2> hiptext.log

# Kill all the subprocesses created in this script if the script itself exits
trap "trap - SIGTERM && kill -- -$$" SIGINT SIGTERM EXIT
