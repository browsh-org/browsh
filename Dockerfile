FROM bitnami/minideb:stretch
RUN install_packages xvfb libgtk-3-0
RUN useradd user
RUN su user
WORKDIR /home/user
ADD ./interfacer/contrib/setup_firefox.sh .
RUN ./setup_firefox.sh
RUN rm ./setup_firefox.sh
ENV PATH="/home/user/firefox:${PATH}"
ADD ./interfacer/contrib/setup_browsh.sh .
RUN ./setup_browsh.sh

CMD ["/home/user/browsh"]

