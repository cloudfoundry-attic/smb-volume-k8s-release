package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"testing"
)

func TestLogForwarder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LogForwarder Suite")
}

var _ = BeforeSuite(func() {
	var err error
	_, err = gexec.Build("code.cloudfoundry.org/volume-services-log-forwarder", "-race")
	Expect(err).NotTo(HaveOccurred())
})
