#!/bin/bash
ffmpeg -f x11grab -s 3200x1600 -r 12 -i :0.0 -vcodec mpeg2video -f mpegts udp://127.0.0.1:1234 >> ffmpeg.log 2>&1 &
sleep 1
# (/home/tombh/.rbenv/versions/2.3.0/bin/ruby ./forward_mouse.rb <&3 &) 3<&0
(./do.bash <&3 &) 3<&0
hiptext -width 95 -font /usr/share/fonts/TTF/DejaVuSansMono.ttf 'udp://127.0.0.1:1234' 2> hiptext.log
trap "trap - SIGTERM && kill -- -$$" SIGINT SIGTERM EXIT
