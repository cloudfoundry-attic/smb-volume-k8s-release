FROM ubuntu:xenial

RUN apt-get update
RUN apt-get install -y --no-install-recommends \
  apt-transport-https \
  ca-certificates \
  curl \
  dnsutils \
  git \
  iproute2 \
  openssh-client \
  software-properties-common \
  jq \
  libpq-dev

# ruby-install dependencies
RUN apt-get install -y \
  wget \
  build-essential \
  libyaml-dev \
  libgdbm-dev \
  libreadline-dev \
  libncurses5-dev \
  libffi-dev

ADD go-version /tmp/go-version
RUN wget https://dl.google.com/go/go$(cat /tmp/go-version).linux-amd64.tar.gz
RUN tar -xvf go$(cat /tmp/go-version).linux-amd64.tar.gz
RUN mv go /usr/local

ENV PATH "$PATH:/usr/local/go/bin"
ENV GOPATH /go
ENV GOROOT=/usr/local/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

RUN apt-get update && apt-get install -y apt-transport-https
RUN curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
RUN echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | tee -a /etc/apt/sources.list.d/kubernetes.list
RUN apt-get update
RUN apt-get install -y kubectl
RUN ln -s /usr/bin/kubectl /usr/local/bin/kubectl

ADD install-ruby.sh /tmp/install-ruby.sh
RUN chmod +x /tmp/install-ruby.sh
RUN /tmp/install-ruby.sh && rm /tmp/install-ruby.sh
ENV PATH /opt/rubies/ruby-2.6.3/bin:$PATH

ADD install-docker.sh /tmp/install-docker.sh
RUN /tmp/install-docker.sh

COPY start-docker.sh /usr/local/bin/start-docker
RUN chmod +x /usr/local/bin/start-docker

RUN wget https://github.com/buildpacks/pack/releases/download/v0.9.0/pack-v0.9.0-linux.tgz
RUN tar xvf pack-v0.9.0-linux.tgz
RUN mv ./pack /usr/local/bin
RUN rm pack-v0.9.0-linux.tgz

RUN wget https://github.com/k14s/kapp/releases/download/v0.21.0/kapp-linux-amd64
RUN mv ./kapp-linux-amd64 /usr/local/bin/kapp
RUN chmod +x /usr/local/bin/kapp

RUN wget https://github.com/k14s/ytt/releases/download/v0.26.0/ytt-linux-amd64
RUN mv ./ytt-linux-amd64 /usr/local/bin/ytt
RUN chmod +x /usr/local/bin/ytt

VOLUME /var/lib/docker
