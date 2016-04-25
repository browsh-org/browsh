FROM alpine

RUN echo "@testing http://nl.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories

# Main dependencies
RUN apk add --no-cache xvfb xdotool@testing xfce4 ffmpeg openssh mosh chromium

# Installing Hiptext, video to text renderer
RUN apk --no-cache add --virtual build-dependencies \
  build-base git freetype-dev jpeg-dev ffmpeg-dev ragel
RUN apk --no-cache add libgflags-dev@testing glog-dev@testing
RUN mkdir -p build \
  && cd build \
  && wget -q -O /etc/apk/keys/andyshinn.rsa.pub https://raw.githubusercontent.com/andyshinn/alpine-pkg-glibc/master/andyshinn.rsa.pub \
  && wget https://github.com/andyshinn/alpine-pkg-glibc/releases/download/2.23-r1/glibc-2.23-r1.apk \
  # && wget https://github.com/andyshinn/alpine-pkg-glibc/releases/download/2.23-r1/glibc-bin-2.23-r1.apk \
  # && wget https://github.com/andyshinn/alpine-pkg-glibc/releases/download/2.23-r1/glibc-i18n-2.23-r1.apk \
  && apk --no-cache add glibc-2.23-r1.apk \
  # && apk add glibc-2.23-r1.apk glibc-bin-2.23-r1.apk glibc-i18n-2.23-r1.apk \
  # && /usr/glibc-compat/bin/localedef -i en_US -f UTF-8 en_US.UTF-8 \
  && git clone https://github.com/tombh/hiptext \
  && cd hiptext \
  && git checkout ffmpeg-updates \
  && make \
  # Alpine's version of `install` doesn't support the `--mode=` format
  && install -m 0755 hiptext /usr/local/bin \
  && cd ../.. && rm -rf build \

  && apk --no-cache del build-dependencies

COPY . /app

# CMD ["run.bash"]
