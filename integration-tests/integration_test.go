package integration_tests_test

import (
	local_k8s_cluster "code.cloudfoundry.org/local-k8s-cluster"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"strings"
)

var _ = Describe("Integration", func() {
	var output string

	BeforeEach(func() {
		var podIP string

		By("deploying a smb server", func() {
			overrides := `{"spec": {"template":  {"spec": {"containers": [{"name": "test-smb1", "command": [ "/sbin/tini", "--", "/usr/bin/samba.sh", "-p","-u","user;pass","-s","user;/export;no;no;no;user","-p","-S" ], "image": "dperson/samba", "securityContext":{"privileged":true}, "ports": [{"containerPort": 139, "protocol": "TCP"}, {"containerPort": 445, "protocol": "TCP"}]}]}}}}`
			local_k8s_cluster.Kubectl("run", "--overrides", overrides, "--image", "dperson/samba", "test-smb1")
			Eventually(func() string {
				podIP = local_k8s_cluster.Kubectl("get", "pods", "-l", "run=test-smb1", "-o", "jsonpath={.items[0].status.podIPs[0].ip}")
				return podIP
			}).Should(Not(Equal("")))
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
			requestPayload := fmt.Sprintf(`{ "service_id": "123", "plan_id": "plan-id", "parameters": {"share": "%s"} }`, podIP)
			request, err := http.NewRequest("PUT", fmt.Sprintf("http://localhost/v2/service_instances/%s", instanceID), strings.NewReader(requestPayload))
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

		Expect(local_k8s_cluster.Kubectl("apply", "-f", "./assets/writer_pod.yaml")).To(ContainSubstring("created"))
		Expect(local_k8s_cluster.Kubectl("apply", "-f", "./assets/reader_pod.yaml")).To(ContainSubstring("created"))

		mountCommand := fmt.Sprintf("mkdir /instance1 && mount -t cifs -o username=user,password=pass //%s/user /instance1 && cat /instance1/foo", podIP)
		output = local_k8s_cluster.Kubectl("exec", "-n", "eirini", "-i", "integration-test-reader", "bash", "-c", mountCommand)
	})

	It("the file contents written by a pod with a pvc (created by the broker) should be written to the smb share", func() {
		Expect(output).To(Equal("hi"))
	})
})
