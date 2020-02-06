package integration_tests_test

import (
	. "github.com/onsi/ginkgo"
	"time"
)

var _ = Describe("Integration", func() {

	It("sleep for a minute", func() {
		time.Sleep(time.Minute)
	})
})
