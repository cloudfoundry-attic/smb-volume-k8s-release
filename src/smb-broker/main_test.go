package main_test

import (
	. "code.cloudfoundry.org/smb-broker"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Main", func() {
	It("hi", func() {
		Expect(Hi()).To(Equal("hi"))
	})
})
