IMAGE_NAME := cfpersi/log-forwarder
THIS_FILE := $(lastword $(MAKEFILE_LIST))

all: install

install:
	go install -v

test:
	GO111MODULE=off go get github.com/onsi/ginkgo/ginkgo && \
	go generate ./... && \
	ginkgo -v -r -keepGoing -trace -randomizeAllSpecs -progress --nodes=1

publish:
	docker login -u $$DOCKER_USER -p $$DOCKER_PASSWORD && \
    pack build cfpersi/log-forwarder --builder cloudfoundry/cnb:bionic --publish

fmt:
	go fmt ./...

tag	:=	"$(shell (git symbolic-ref -q --short HEAD || git describe --tags --exact-match) | sed 's/master/latest/')"
kapp: SHELL=/bin/bash
kapp:
	kubectl get namespace cf-workloads || kubectl create namespace cf-workloads
	kubectl get namespace cf-smb || kubectl create namespace cf-smb
	kapp deploy -y -a volume-services-log-forwarder -f <(ytt -f ytt/)

.PHONY: test kapp