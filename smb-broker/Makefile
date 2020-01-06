IMAGE_NAME      :=      cfpersi/smb-broker-k8s

all: install

install:
	go install -v

test:
	GO111MODULE=off go get github.com/onsi/ginkgo/ginkgo
	GO111MODULE=on ginkgo -v -r -keepGoing -trace -randomizeAllSpecs -progress --nodes=1

fmt:
	go fmt ./... -v

image:
	docker build -t cfpersi/smb-broker .
	docker tag cfpersi/smb-broker cfpersi/smb-broker:dev
	docker push cfpersi/smb-broker:dev

helm:
	helm status smb-dev-broker | grep 'STATUS: DEPLOYED' && helm upgrade --recreate-pods smb-dev-broker ./helm || helm install --name smb-dev-broker ./helm

helm-dev:
	helm status smb-dev-broker | grep 'STATUS: DEPLOYED' && helm upgrade --recreate-pods --set image.tag=dev smb-dev-broker ./helm || helm install --name smb-dev-broker --set image.tag=dev ./helm

helm_delete:
	helm delete --purge smb-dev-broker

fly:
	fly -t persi execute -c /Users/pivotal/workspace/smb-volume-k8s-release/smb-broker/ci/unit-tests.yml -i smb-volume-k8s-release=/Users/pivotal/workspace/smb-volume-k8s-release -p

.PHONY: install test fmt image release helm helm_delete helm-dev fly
