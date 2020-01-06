IMAGE_NAME      :=      cfpersi/smb-broker-k8s

all: install

install:
	go install -v

test:
	GO111MODULE=off go get github.com/onsi/ginkgo/ginkgo
	GO111MODULE=on ginkgo -v -r -keepGoing -trace -randomizeAllSpecs -progress --nodes=1

fmt:
	go fmt ./... -v

running	:=	"$(shell docker inspect -f '{{.State.Running}}' "kind-registry" 2>/dev/null || true)"
image-local-registry: SHELL:=/bin/bash
image-local-registry:
	[ $(running) != "true" ] && docker run \
	    -d --restart=always -p "5000:5000" --name "kind-registry" \
	    registry:2 || true
	docker build --no-cache -t cfpersi/smb-broker .
	docker tag cfpersi/smb-broker localhost:5000/cfpersi/smb-broker:local-test
	docker push localhost:5000/cfpersi/smb-broker:local-test

helm:
	helm status smb-dev-broker | grep 'STATUS: DEPLOYED' && helm upgrade --recreate-pods smb-dev-broker ./helm || helm install --name smb-dev-broker ./helm

helm-dev:
	helm status smb-dev-broker | grep 'STATUS: DEPLOYED' && helm upgrade --recreate-pods --set image.tag=dev smb-dev-broker ./helm || helm install --name smb-dev-broker --set image.tag=dev ./helm

helm_delete:
	helm delete --purge smb-dev-broker

fly:
	fly -t persi execute -c /Users/pivotal/workspace/smb-volume-k8s-release/smb-broker/ci/unit-tests.yml -i smb-volume-k8s-release=/Users/pivotal/workspace/smb-volume-k8s-release -p

.PHONY: install test fmt image-local-registry release helm helm_delete helm-dev fly
