FROM alpine
RUN apk add xvfb xdotool xfce4 ffmpeg chromium
RUN rm -rf /var/cache/apk/*
