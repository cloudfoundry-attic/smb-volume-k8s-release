package identityserver_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIdentityserver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Identityserver Suite")
}
