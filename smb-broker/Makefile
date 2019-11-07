VERSION         :=      $(shell cat ./VERSION)
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

# Goreleaser is a tool that allows anyone to integrate a binary releasing
# process to their pipelines.
release:
	git tag -a $(VERSION) -m "Release" || true
	git push origin $(VERSION)
	goreleaser --rm-dist

.PHONY: install test fmt image release helm helm_delete