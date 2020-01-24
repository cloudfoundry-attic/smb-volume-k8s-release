package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSmbCsiDriver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SmbCsiDriver Suite")
}
