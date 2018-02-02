FROM bitnami/minideb:stretch

RUN install_packages xvfb libgtk-3-0 curl ca-certificates bzip2 libdbus-glib-1-2
RUN useradd -m user
RUN su user
ENV HOME=/home/user
WORKDIR $HOME

# These are needed to detect versions
ADD .travis.yml .
ADD ./webext/manifest.json .

# Setup Firefox
ENV PATH="/home/user/bin/firefox:${PATH}"
ADD ./interfacer/contrib/setup_firefox.sh .
RUN ./setup_firefox.sh
RUN rm ./setup_firefox.sh && rm .travis.yml

# Setup Browsh
ADD ./interfacer/contrib/setup_browsh.sh .
RUN ./setup_browsh.sh
RUN rm ./setup_browsh.sh && rm manifest.json

CMD ["/home/user/browsh"]

