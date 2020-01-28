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

COPY --from=golang:1 /usr/local/go /usr/local/go
ENV GOROOT=/usr/local/go PATH=/usr/local/go/bin:$PATH

RUN apt-get update && apt-get install -y apt-transport-https
RUN curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
RUN echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | tee -a /etc/apt/sources.list.d/kubernetes.list
RUN apt-get update
RUN apt-get install -y kubectl

RUN wget https://get.helm.sh/helm-v3.0.2-linux-amd64.tar.gz
RUN tar -zxvf helm-v3.0.2-linux-amd64.tar.gz
RUN mv linux-amd64/helm /usr/local/bin/helm

ADD install-ruby.sh /tmp/install-ruby.sh
RUN chmod +x /tmp/install-ruby.sh
RUN /tmp/install-ruby.sh && rm /tmp/install-ruby.sh
ENV PATH /opt/rubies/ruby-2.6.3/bin:$PATH

ADD install-docker.sh /tmp/install-docker.sh
RUN /tmp/install-docker.sh

COPY start-docker.sh /usr/local/bin/start-docker
RUN chmod +x /usr/local/bin/start-docker

VOLUME /var/lib/docker
