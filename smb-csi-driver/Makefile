THIS_FILE := $(lastword $(MAKEFILE_LIST))

vet:
	go vet .

fakes:
	go generate ./...

build:
	go build .

running	:=	"$(shell docker inspect -f '{{.State.Running}}' "kind-registry" 2>/dev/null || true)"
image-local-registry: SHELL:=/bin/bash
image-local-registry:
	[ $(running) != "true" ] && docker run \
	    --ip 172.17.0.2 -d --restart=always -p "5000:5000" --name "kind-registry" \
	    registry:2 || true
	pack build localhost:5000/cfpersi/smb-plugin:local-test --builder cloudfoundry/cnb:bionic --buildpack paketo-buildpacks/go-compiler --run-image cfpersi/cnb-cifs-run-stack --publish

start-docker:
	start-docker &
	echo 'until docker info; do sleep 5; done' >/usr/local/bin/wait_for_docker
	chmod +x /usr/local/bin/wait_for_docker
	timeout 300 wait_for_docker

kill-docker:
	pkill dockerd

test:
	@$(MAKE) -f $(THIS_FILE) vet
	@$(MAKE) -f $(THIS_FILE) fakes
	@$(MAKE) -f $(THIS_FILE) build
	go get github.com/onsi/ginkgo/ginkgo
	cd identityserver && ginkgo -race .
	cd nodeserver && ginkgo -race .

e2e: SHELL:=/bin/bash
e2e:
	@$(MAKE) -f $(THIS_FILE) image-local-registry
	go get github.com/onsi/ginkgo/ginkgo
	ginkgo -r -focus "CSI Volumes|CSIDriver"

fly:
	fly -t persi execute --tag=kind  -p -c ~/workspace/smb-volume-k8s-release/smb-csi-driver/ci/unit-tests.yml -i smb-volume-k8s-release=$$HOME/workspace/smb-volume-k8s-release

fly-e2e:
	fly -t persi execute --tag=kind  -p -c ~/workspace/smb-volume-k8s-release/smb-csi-driver/ci/e2e-tests.yml -i smb-volume-k8s-release=$$HOME/workspace/smb-volume-k8s-release

tag	:=	"$(shell (git symbolic-ref -q --short HEAD || git describe --tags --exact-match) | sed 's/master/latest/')"
kapp: SHELL=/bin/bash
kapp:
	kubectl get namespace cf-smb || kubectl create namespace cf-smb
	kapp deploy -y -a smb-csi-driver -f <(ytt -f ytt/base -v image.tag=$(tag))

.PHONY: test fly fly-e2e image-local-registry