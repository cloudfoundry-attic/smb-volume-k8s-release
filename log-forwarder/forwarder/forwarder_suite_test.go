package forwarder_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestForwarder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Forwarder Suite")
}
