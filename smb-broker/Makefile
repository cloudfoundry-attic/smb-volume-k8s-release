IMAGE_NAME      :=      cfpersi/smb-broker-k8s

all: install

install:
	go install -v

test:
	GO111MODULE=off go get github.com/onsi/ginkgo/ginkgo
	ginkgo -r -keepGoing -p -trace -randomizeAllSpecs -progress

fmt:
	go fmt ./... -v

image:
	docker build -t cfpersi/smb-broker .

helm:
	helm status smb-dev-broker | grep 'STATUS: DEPLOYED' && helm upgrade smb-dev-broker ./helm || helm install --name smb-dev-broker ./helm

helm_delete:
	helm delete --purge smb-dev-broker

.PHONY: install test fmt image release helm helm_delete