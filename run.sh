#!/bin/sh

export LC_ALL=C
export LANG=C

export DESKTOP_WIDTH='1600'
export TTY_WIDTH=$(( $(stty size | cut -d' ' -f2) - 1))
export TTY_HEIGHT=$(( $(stty size | cut -d' ' -f1) - 1))
# Hiptext uses a row to represent twice as much as a column in order
# to more faithfully project the aspect ratio of the image/video.
ratio=$(echo "scale=5; $TTY_HEIGHT * 2 / $TTY_WIDTH" | bc)
height_float=$(echo "scale=5; $ratio*$DESKTOP_WIDTH" | bc)
export DESKTOP_HEIGHT=$(printf "%.0f\n" "$height_float")

export DISPLAY=:0
DESKTOP_RES="$DESKTOP_WIDTH"x"$DESKTOP_HEIGHT"
UDP_URI='udp://127.0.0.1:1234'

# Create an X desktop in memory without actually displaying it on a real screen.
# Double the width to make room for the xzoom window, which is actually what
# ffmpeg will stream;
# ---------------------------------
# |              |                |
# |  desktop     |   xzoom win    |
# |  here        |   here mirrors |
# |              |   desktop      |
# |              |                |
# ---------------------------------
# So xzoom mirrors the desktop and ffmpeg streams the xzoom window.
Xvfb :0 -screen 0 "$(($DESKTOP_WIDTH * 2))"x"$DESKTOP_HEIGHT"x16 > ./logs/xvfb.log 2>&1 &

# TODO: detect X start rather than sleep
sleep 1

/usr/bin/firefox >> ./logs/xvfb.log 2>&1 &

# Convert the X framebuffer desktop into a video stream, but only stream the
# right hand side where the xzoom window is.
# TODO: Can latency be reduced further? Can flicker be reduced, in order to reduce bandwidth?
ffmpeg \
  -f x11grab \
  -s $DESKTOP_RES \
  -r 12 \
  -i :0.0+$DESKTOP_WIDTH \
  -vcodec mpeg2video \
  -f mpegts \
  $UDP_URI \
  > ./logs/ffmpeg.log 2>&1 &

# The above ffmpeg can take a while to open the UDP stream, so wait a little
# TODO: detect the stream's presence rather than sleep
sleep 1

# Intercept STDIN (mouse and keypresses) and forward to the X framebuffer via xdotool
(
  # Kill all the processes in this script when the interfacer exits
  trap '
    trap - EXIT
    killall Xvfb firefox ffmpeg hiptext
    exit
  ' EXIT INT TERM

  ./interfacer/interfacer <&3 > ./logs/interfacer.log 2>&1
) 3<&0 &

# Hiptext renders images and videos into text characters displayable in a terminal.
# It complains unless you specify the exact path to the font, seems like a bug to me.
# TODO: support dynamic sizing
hiptext \
  -font /usr/share/fonts/ttf-dejavu/DejaVuSansMono.ttf \
  --xterm256unicode \
  -bgprint=true \
  -fast \
  $UDP_URI \
  2> ./logs/hiptext.log
