module code.cloudfoundry.org/smb-broker

go 1.12

require (
	code.cloudfoundry.org/lager v2.0.0+incompatible
	code.cloudfoundry.org/local-k8s-cluster v0.0.0
	github.com/drewolson/testflight v1.0.0 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/maxbrunsfeld/counterfeiter/v6 v6.2.2
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/pivotal-cf/brokerapi v6.4.2+incompatible
	github.com/pkg/errors v0.9.0
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
	sigs.k8s.io/kind v0.7.0
)

replace code.cloudfoundry.org/local-k8s-cluster => ../local-k8s-cluster
