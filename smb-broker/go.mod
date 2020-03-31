module code.cloudfoundry.org/smb-broker

go 1.13

require (
	code.cloudfoundry.org/lager v2.0.0+incompatible
	code.cloudfoundry.org/smb-volume-k8s-local-cluster v1.0.1-0.20200326224655-f9b595631b53
	github.com/drewolson/testflight v1.0.0 // indirect
	github.com/gorilla/mux v1.7.4
	github.com/maxbrunsfeld/counterfeiter/v6 v6.2.3
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/pivotal-cf/brokerapi v6.4.2+incompatible
	github.com/pkg/errors v0.9.1
	k8s.io/api v0.18.0
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v0.18.0
)
