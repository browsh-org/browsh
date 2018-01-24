FROM phusion/baseimage

RUN apt-get update
RUN apt-get -y install xvfb libgtk-3-0

RUN useradd user
RUN su user
WORKDIR /home/user
RUN curl -L -o firefox.tar.bz2 https://ftp.mozilla.org/pub/firefox/releases/58.0b16/linux-x86_64/en-US/firefox-58.0b16.tar.bz2
RUN apt-get -y install bzip2
RUN bzip2 -d firefox.tar.bz2
RUN tar xf firefox.tar
ENV PATH="/home/user/firefox:${PATH}"
ADD interfacer/browsh .

CMD ["/home/user/browsh"]

