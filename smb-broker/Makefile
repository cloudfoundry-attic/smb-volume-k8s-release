IMAGE_NAME      :=      cfpersi/smb-broker-k8s
THIS_FILE := $(lastword $(MAKEFILE_LIST))

all: install

install:
	go install -v

test:
	@$(MAKE) -f $(THIS_FILE) image-local-registry
	@$(MAKE) -f $(THIS_FILE) push-local-image-source
	ulimit -n 10000 && ulimit -a && GO111MODULE=off go get github.com/onsi/ginkgo/ginkgo && GO111MODULE=on ginkgo -v -r -keepGoing -trace -randomizeAllSpecs -progress --nodes=1

fmt:
	go fmt ./...

running	:=	"$(shell docker inspect -f '{{.State.Running}}' "kind-registry" 2>/dev/null || true)"
image-local-registry: SHELL:=/bin/bash
image-local-registry:
	[ $(running) != "true" ] && docker run \
	    --ip 172.17.0.2 -d --restart=always -p "5000:5000" --name "kind-registry" \
	    registry:2 || true

push-local-image-source:
	docker login localhost:5000 && \
	docker build --no-cache -f ./test/Dockerfile -t cfpersi/smb-broker-source /Users/pivotal/workspace/smb-volume-k8s-release/smb-broker/ && \
	docker tag cfpersi/smb-broker-source localhost:5000/cfpersi/smb-broker-source:latest && \
	docker push localhost:5000/cfpersi/smb-broker-source:latest

helm:
	helm version | grep 'v3.' || ( echo 'ERROR: requires helm 3!!!!'; exit 1 )
	helm status smb-broker | grep 'STATUS: deployed' && helm upgrade --recreate-pods smb-broker --set ingress.enabled=true ./helm || helm install smb-broker --set ingress.enabled=true ./helm

helm-dev:
	helm version | grep 'v3.0' || ( echo 'ERROR: requires helm 3!!!!'; exit 1 )
	helm status smb-dev-broker | grep 'STATUS: deployed' && helm upgrade --recreate-pods --set ingress.enabled=true --set image.tag=dev smb-dev-broker ./helm || helm install smb-dev-broker --set ingress.enabled=true --set image.tag=dev ./helm

helm_dev_delete:
	helm delete smb-dev-broker

helm_delete:
	helm delete smb-broker

helm_scf:
	helm version | grep 'v3.0' || ( echo 'ERROR: requires helm 3!!!!'; exit 1 )
	helm status smb-broker | grep 'STATUS: deployed' && helm upgrade --recreate-pods smb-broker --set targetNamespace=eirini --set ingress.enabled=true ./helm || helm install smb-broker --set ingress.enabled=true --set targetNamespace=eirini ./helm

fly:
	fly -t persi execute -c $$HOME/workspace/smb-volume-k8s-release/smb-broker/ci/unit-tests.yml -i smb-volume-k8s-release=$$HOME/workspace/smb-volume-k8s-release -p

.PHONY: test fly image-local-registry push-local-image-source helm
