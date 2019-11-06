VERSION         :=      $(shell cat ./VERSION)
IMAGE_NAME      :=      cfpersi/smb-broker-k8s

all: install

install:
	go install -v

test:
	ginkgo -r -keepGoing -p -trace -randomizeAllSpecs -progress

fmt:
	go fmt ./... -v

image:
	docker build -t cfpersi/smb-broker .


# Goreleaser is a tool that allows anyone to integrate a binary releasing
# process to their pipelines.
release:
	git tag -a $(VERSION) -m "Release" || true
	git push origin $(VERSION)
	goreleaser --rm-dist

.PHONY: install test fmt image release