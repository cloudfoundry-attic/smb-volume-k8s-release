IMAGE_NAME := cfpersi/log-forwarder
THIS_FILE := $(lastword $(MAKEFILE_LIST))

all: install

install:
	go install -v

test:
	@$(MAKE) -f $(THIS_FILE) image-local-registry
	ulimit -n 10000 && ulimit -a && \
	GO111MODULE=off go get github.com/onsi/ginkgo/ginkgo && \
	GO111MODULE=on go generate ./... && \
	GO111MODULE=on ginkgo -v -r -keepGoing -trace -randomizeAllSpecs -progress --nodes=1

fmt:
	go fmt ./...

running	:=	"$(shell docker inspect -f '{{.State.Running}}' "kind-registry" 2>/dev/null || true)"
image-local-registry: SHELL:=/bin/bash
image-local-registry:
	[ $(running) != "true" ] && docker run \
	    --ip 172.17.0.2 -d --restart=always -p "5000:5000" --name "kind-registry" \
	    registry:2 || true
	pack build localhost:5000/$(IMAGE_NAME):local-test --builder cloudfoundry/cnb:bionic --publish

tag	:=	"$(shell (git symbolic-ref -q --short HEAD || git describe --tags --exact-match) | sed 's/master/latest/')"
kapp: SHELL=/bin/bash
kapp:
	kubectl get namespace cf-workloads || kubectl create namespace cf-workloads
	kubectl get namespace cf-smb || kubectl create namespace cf-smb
	kapp deploy -y -a volume-services-log-forwarder -f <(ytt -f ytt/ -v image.tag=$(tag))

fly:
	fly -t persi execute --tag=kind -c $$HOME/workspace/smb-volume-k8s-release/log-forwarder/ci/unit-tests.yml -i smb-volume-k8s-release=$$HOME/workspace/smb-volume-k8s-release -p

.PHONY: test fly image-local-registry kapp
