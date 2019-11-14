package main_test

import (
	"fmt"
	"github.com/onsi/gomega/gexec"
	"io"
	"io/ioutil"
	"path"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSmbBroker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SmbBroker Suite")
}

var smbBrokerCompiledPath string

var _ = BeforeSuite(func() {
	var err error
	smbBrokerCompiledPath, err = gexec.Build("code.cloudfoundry.org/smb-broker", "-mod=vendor")
	Expect(err).NotTo(HaveOccurred())
})


func fixture(name string) string {
	filePath := path.Join("fixtures", name)
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(fmt.Sprintf("Could not read fixture: %s", name))
	}

	return string(contents)
}

func assertHttpResponseContainsSubstring(body io.Reader, expected string) {
	bytes, err := ioutil.ReadAll(body)
	Expect(err).NotTo(HaveOccurred())
	Expect(string(bytes)).Should(ContainSubstring(expected))
}