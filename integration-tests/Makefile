test:
	go get github.com/onsi/ginkgo/ginkgo
	make --directory=../smb-broker image-local-registry
	make --directory=../smb-csi-driver image-local-registry
	~/go/bin/ginkgo -race .

fly:
	# fly -t persi execute -p -c ~/workspace/smb-volume-k8s-release/smb-csi-driver/ci/integration-tests.yml -i smb-volume-k8s-release=/Users/pivotal/workspace/smb-volume-k8s-release