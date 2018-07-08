FROM bitnami/minideb:stretch

RUN install_packages xvfb libgtk-3-0 curl ca-certificates bzip2 libdbus-glib-1-2 procps

# Logging client for Google's Stackdriver logging service.
# NB Not used by default. Only used by the Browsh as a Service platform on the
#    anonymous accounts.
RUN curl -L -o /usr/local/bin/gcloud_logger https://github.com/tombh/gcloud_pipe_logger/releases/download/v0.0.5/gcloud_pipe_logger_0.0.5_linux_amd64
RUN chmod a+x /usr/local/bin/gcloud_logger

RUN curl -o /etc/hosts https://raw.githubusercontent.com/StevenBlack/hosts/master/alternates/fakenews-gambling-porn-social/hosts

RUN useradd -m user --home /app
USER user
ENV HOME=/app
WORKDIR /app

# These are needed to detect versions
ADD .travis.yml .
ADD ./webext/manifest.json .

# Setup Firefox
ENV PATH="/app/bin/firefox:${PATH}"
ADD ./interfacer/contrib/setup_firefox.sh .
RUN ./setup_firefox.sh
RUN rm ./setup_firefox.sh && rm .travis.yml

# Setup Browsh
ADD ./interfacer/contrib/setup_browsh.sh .
RUN ./setup_browsh.sh
# Firefox behaves quite differently to normal on its first run, so by getting
# that over and done with here when there's no user to be dissapointed means
# that all future runs will be consistent.
RUN TERM=xterm script \
      --return \
      -c "/app/browsh" \
      /dev/null \
      >/dev/null & \
      sleep 10
RUN rm ./setup_browsh.sh && rm manifest.json

CMD ["/app/browsh"]

