FROM alpine
COPY . /app

RUN echo "http://mirror1.hs-esslingen.de/pub/Mirrors/alpine/v3.3/main" > /etc/apk/repositories
RUN echo "http://mirror1.hs-esslingen.de/pub/Mirrors/alpine/v3.3/community" >> /etc/apk/repositories
RUN echo "@testing http://mirror1.hs-esslingen.de/pub/Mirrors/alpine/edge/testing" >> /etc/apk/repositories

# Main dependencies
RUN apk add --no-cache bc xvfb ttf-dejavu xdotool@testing ffmpeg openssh mosh firefox dbus

# Installing Hiptext, video to text renderer and our own interfacer.go
# Keep this all in one RUN command so that the resulting Docker image is smaller.
RUN apk --no-cache add --virtual build-dependencies \
  build-base git go freetype-dev jpeg-dev ffmpeg-dev ragel libx11-dev libxt-dev libxext-dev \
  && apk --no-cache add libgflags-dev@testing glog-dev@testing \
  && mkdir -p build \
  && cd build \

  # Need glibc for locale support, not that any UTF8 locales work :/
  # This PR seems to be the most relevant: https://github.com/andyshinn/alpine-pkg-glibc/issues/13
  # Note that we're currently having to use this hack in hiptext because of alpine's poor locale support:
  # https://github.com/tombh/hiptext/commit/bc502af5f6e3b622a9b53d1ffb9a40e74d968ae3
  && wget -q -O /etc/apk/keys/andyshinn.rsa.pub https://raw.githubusercontent.com/andyshinn/alpine-pkg-glibc/master/andyshinn.rsa.pub \
  && wget https://github.com/andyshinn/alpine-pkg-glibc/releases/download/2.23-r1/glibc-2.23-r1.apk \
  && apk --no-cache add glibc-2.23-r1.apk \

  # Currently need to use a patched vesion of hiptext that supports video streams and ffmpeg v3
  # Watch: https://github.com/jart/hiptext/pull/27
  && git clone https://github.com/tombh/hiptext \
  && cd hiptext \
  && git checkout ffmpeg-updates-and-unicode-hack \
  && make \
  # Alpine's version of `install` doesn't support the `--mode=` format
  && install -m 0755 hiptext /usr/local/bin \
  && cd ../.. && rm -rf build \

  # Build the interfacer.go/xzoom code
  && export GOPATH=/go && export GOBIN=/app/interfacer \
  && cd /app/interfacer && go get && go build \

  && apk --no-cache del build-dependencies

# Generate host keys
RUN ssh-keygen -A

RUN sed -i 's/#Port 22/Port 7777/' /etc/ssh/sshd_config

RUN mkdir -p /app/logs

WORKDIR /app
CMD ["/usr/sbin/sshd", "-D"]
