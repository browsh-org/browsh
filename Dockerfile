FROM bitnami/minideb:stretch as build

RUN install_packages \
      curl \
      ca-certificates \
      git \
      autoconf \
      automake \
      g++ \
      protobuf-compiler \
      zlib1g-dev \
      libncurses5-dev \
      libssl-dev \
      pkg-config \
      libprotobuf-dev \
      make

# Install Golang
ENV GOROOT=/go
ENV GOPATH=/go-home
ENV PATH=$GOROOT/bin:$GOPATH/bin:$PATH
RUN curl -L -o go.tar.gz https://dl.google.com/go/go1.9.2.linux-amd64.tar.gz
RUN mkdir -p $GOPATH/bin
RUN tar -C / -xzf go.tar.gz

ENV BASE=$GOPATH/src/browsh/interfacer
WORKDIR $BASE
ADD interfacer $BASE

# Build Browsh
RUN $BASE/contrib/build_browsh.sh


###########################
# Actual final Docker image
###########################
FROM bitnami/minideb:stretch

ENV HOME=/app
WORKDIR /app

COPY --from=build /go-home/src/browsh/interfacer/browsh /app/browsh

RUN install_packages \
      xvfb \
      libgtk-3-0 \
      curl \
      ca-certificates \
      bzip2 \
      libdbus-glib-1-2 \
      procps

# Block ads, etc. This includes porn just because this image is also used on the
# public SSH demo: `ssh brow.sh`.
RUN curl -o /etc/hosts https://raw.githubusercontent.com/StevenBlack/hosts/master/alternates/fakenews-gambling-porn-social/hosts

# Don't use root
RUN useradd -m user --home /app
RUN chown user:user /app
USER user

# Setup Firefox
ENV PATH="${HOME}/bin/firefox:${PATH}"
ADD .travis.yml .
ADD interfacer/contrib/setup_firefox.sh .
RUN ./setup_firefox.sh
RUN rm setup_firefox.sh && rm .travis.yml

# Firefox behaves quite differently to normal on its first run, so by getting
# that over and done with here when there's no user to be dissapointed means
# that all future runs will be consistent.
RUN TERM=xterm script \
  --return \
  -c "/app/browsh" \
  /dev/null \
  >/dev/null & \
  sleep 10

CMD ["/app/browsh"]

