FROM bitnami/minideb:stretch
RUN install_packages xvfb libgtk-3-0
RUN useradd user
RUN su user
WORKDIR /home/user
ADD ./interfacer/contrib/setup_firefox.sh .
ADD .travis.yml .
RUN ./setup_firefox.sh
RUN rm ./setup_firefox.sh && rm .travis.yml
ENV PATH="/home/user/firefox:${PATH}"
ADD ./webpack/manifest.json .
ADD ./interfacer/contrib/setup_browsh.sh .
RUN ./setup_browsh.sh
RUN rm ./setup_browsh.sh && rm manifest.json

CMD ["/home/user/browsh"]

