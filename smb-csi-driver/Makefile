build:
	go build .

running	:=	"$(shell docker inspect -f '{{.State.Running}}' "kind-registry" 2>/dev/null || true)"
image-local-registry: SHELL:=/bin/bash
image-local-registry:
	[ $(running) != "true" ] && docker run \
	    -d --restart=always -p "5000:5000" --name "kind-registry" \
	    registry:2 || true
	docker build --no-cache -t cfpersi/smb-plugin .
	docker tag cfpersi/smb-plugin localhost:5000/cfpersi/smb-plugin:local-test
	docker push localhost:5000/cfpersi/smb-plugin:local-test

start-docker:
	start-docker &
	until docker info; do sleep 1; done

kill-docker:
	pkill dockerd

test:
	go get github.com/onsi/ginkgo/ginkgo
	~/go/bin/ginkgo -flakeAttempts 3 -race -focus "(Identity|Node) Service"

fly:
	fly -t persi execute -p -c ~/workspace/smb-volume-k8s-release/smb-csi-driver/ci/integration-tests.yml -i smb-volume-k8s-release=/Users/pivotal/workspace/smb-volume-k8s-release

.PHONY: build test image-local-registry
