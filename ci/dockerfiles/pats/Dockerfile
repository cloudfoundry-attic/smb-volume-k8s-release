FROM google/cloud-sdk

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update
RUN apt-get install -y jq wget curl
RUN apt-get install -y awscli
RUN apt-get install -y python3-pip
RUN pip3 install yq

RUN wget https://github.com/k14s/ytt/releases/download/v0.26.0/ytt-linux-amd64
RUN mv ytt-linux-amd64 /usr/local/bin/ytt
RUN chmod +x /usr/local/bin/ytt

RUN wget https://github.com/k14s/kapp/releases/download/v0.26.0/kapp-linux-amd64
RUN mv kapp-linux-amd64 /usr/local/bin/kapp
RUN chmod +x /usr/local/bin/kapp

RUN wget https://github.com/cloudfoundry/bosh-cli/releases/download/v6.2.1/bosh-cli-6.2.1-linux-amd64
RUN mv bosh-cli-6.2.1-linux-amd64 /usr/local/bin/bosh
RUN chmod +x /usr/local/bin/bosh

