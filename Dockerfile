FROM alpine

RUN echo "@testing http://nl.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories

# Main dependencies
RUN apk add --no-cache bc xvfb ttf-dejavu xdotool@testing ffmpeg openssh mosh firefox dbus

# Generate host keys
RUN ssh-keygen -A

# Installing Hiptext, video to text renderer and our own interfacer.go
RUN apk --no-cache add --virtual build-dependencies \
  build-base git go freetype-dev jpeg-dev ffmpeg-dev ragel libx11-dev libxt-dev libxext-dev
RUN apk --no-cache add libgflags-dev@testing glog-dev@testing
RUN mkdir -p build \
  && cd build \

  # Need glibc for locale support
  && wget -q -O /etc/apk/keys/andyshinn.rsa.pub https://raw.githubusercontent.com/andyshinn/alpine-pkg-glibc/master/andyshinn.rsa.pub \
  && wget https://github.com/andyshinn/alpine-pkg-glibc/releases/download/2.23-r1/glibc-2.23-r1.apk \
  && apk --no-cache add glibc-2.23-r1.apk \

  # Currently need to use a patched vesion of hiptext that supports video streams and ffmpeg v3
  # Watch: https://github.com/jart/hiptext/pull/27
  && git clone https://github.com/tombh/hiptext \
  && cd hiptext \
  && git checkout ffmpeg-updates \
  && make \
  # Alpine's version of `install` doesn't support the `--mode=` format
  && install -m 0755 hiptext /usr/local/bin \
  && cd ../.. && rm -rf build

COPY . /app

RUN export GOPATH=/go && export GOBIN=/app/interfacer/ && \
    cd /app/interfacer && go get && go build

RUN mkdir -p /app/logs

RUN apk --no-cache del build-dependencies

RUN sed -i 's/#Port 22/Port 7777/' /etc/ssh/sshd_config

WORKDIR /app
CMD ["/usr/sbin/sshd", "-D"]
