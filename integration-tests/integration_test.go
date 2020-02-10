package integration_tests_test

import (
	local_k8s_cluster "code.cloudfoundry.org/local-k8s-cluster"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"strings"
	"time"
)

var _ = Describe("Integration", func() {

	BeforeEach(func() {
		By("deploying a smb server", func() {
			overrides := `{"spec": {"template":  {"spec": {"containers": [{"name": "test-smb1", "command": [ "/sbin/tini", "--", "/usr/bin/samba.sh", "-p","-u","user;pass","-s","user;/export;no;no;no;user","-p","-S" ], "image": "dperson/samba", "securityContext":{"privileged":true}, "ports": [{"containerPort": 139, "protocol": "TCP"}, {"containerPort": 445, "protocol": "TCP"}]}]}}}}`
			local_k8s_cluster.Kubectl("run", "--overrides", overrides, "--image", "dperson/samba", "test-smb1")
		})

		var resp *http.Response

		Eventually(func() string {
			resp, _ = http.DefaultClient.Get("http://localhost/v2/catalog")
			if resp == nil {
				return ""
			}
			return resp.Status
		}).Should(Equal("200 OK"))

		instanceID := "instance1"
		bindingID := "binding1"

		Eventually(func() string {
			request, err := http.NewRequest("PUT", fmt.Sprintf("http://localhost/v2/service_instances/%s", instanceID), strings.NewReader(`{ "service_id": "123", "plan_id": "plan-id" }`))
			Expect(err).NotTo(HaveOccurred())

			resp, _ = http.DefaultClient.Do(request)
			if resp == nil {
				return ""
			}
			return resp.Status
		}).Should(Equal("201 Created"))

		Eventually(func() string {
			request, err := http.NewRequest("PUT", fmt.Sprintf("http://localhost/v2/service_instances/%s/service_bindings/%s", instanceID, bindingID),
				strings.NewReader(`{"service_id": "123", "plan_id": "plan_id", "bind_resource": {"app_guid": "123"}}`))

			Expect(err).NotTo(HaveOccurred())
			resp, _ = http.DefaultClient.Do(request)
			if resp == nil {
				return ""
			}
			return resp.Status
		}).Should(Equal("201 Created"))

		// curl smb broker create-service
		// curl smb broker bind-service

		// dig the pvc name out of the bind-service repsonse

		// kubectl apply a pod referencing the pvc

		// exec into the pod, write a file
		// read the file from the smb share->assert the contents are correct
	})

	It("sleep for a minute", func() {
		time.Sleep(time.Minute)
	})
})
