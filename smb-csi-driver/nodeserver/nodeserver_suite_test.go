package nodeserver_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestNodeserver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Nodeserver Suite")
}
