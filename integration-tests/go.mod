module code.cloudfoundry.org/smb-k8s-integration-tests

go 1.12

require (
	code.cloudfoundry.org/local-k8s-cluster v0.0.0
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
)

replace code.cloudfoundry.org/local-k8s-cluster => ../local-k8s-cluster
