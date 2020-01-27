build:
	go build .

test:
	go get github.com/onsi/ginkgo/ginkgo
	ginkgo -flakeAttempts 3 -race -focus "(Identity|Node) Service"

fly:
	fly -t persi execute -c ~/workspace/smb-volume-k8s-release/smb-csi-driver/ci/integration-tests.yml -i smb-volume-k8s-release=/Users/pivotal/workspace/smb-volume-k8s-release

.PHONY: build test
