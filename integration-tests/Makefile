test:
	go get github.com/onsi/ginkgo/ginkgo
	make --directory=../smb-broker image-local-registry
	make --directory=../smb-csi-driver image-local-registry
	~/go/bin/ginkgo -v -r -keepGoing -trace -randomizeAllSpecs -progress --nodes=1 .

start-docker:
	start-docker &
	echo 'until docker info; do sleep 5; done' >/usr/local/bin/wait_for_docker
	chmod +x /usr/local/bin/wait_for_docker
	timeout 300 wait_for_docker

kill-docker:
	pkill dockerd