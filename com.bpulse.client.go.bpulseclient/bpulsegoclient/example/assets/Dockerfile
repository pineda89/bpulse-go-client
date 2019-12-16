FROM golang:1.4.2
MAINTAINER TheSkyInFlames <theskyinflames@gmail.com>

VOLUME /var/lib/docker

RUN mkdir /assets
RUN mkdir /conf

ENV GOROOT_BOOTSTRAP /usr/src/go

ENV GO_VERSION go1.5.1

ENV BPULSEID bpulse_serhs_TEST
ENV BPULSE_HOST localhost
ENV BPULSE PORT 21000
ENV BPULSE_USER user
ENV BPULSE_PWD pwd
ENV BPULSE_BUILD_TAGS debug

ENV GO_COMMON_LIB_CONFIGURATION_FILE=/conf/BPulseGoClient.ini
ENV GO_COMMON_LIB_EVENTS_FILE=/conf/EventsConfiguration.json

ENV EXAMPLE_LOOP_DELAY_IN_MS 500

ENV GVM_ROOT /root/.gvm

ENV CODE_BRANCH master
ENV REPO_NAME bpulse-go-client
ENV REPO_NAME_PREFIX github.com/pineda89
ENV REPO https://github.com/pineda89/bpulse-go-client.git

RUN rm /bin/sh && ln -s /bin/bash /bin/sh
RUN apt-get update -y

RUN mkdir -p /gobin

WORKDIR /assets
ADD install_go_version.sh /assets/
RUN chmod +x install_go_version.sh; sync; ./install_go_version.sh

ADD build_example.sh /assets/
RUN chmod +x build_example.sh && sync && ./build_example.sh

ADD run_example.sh /assets/
CMD chmod +x ./run_example.sh && sync && ls -la && ./run_example.sh
