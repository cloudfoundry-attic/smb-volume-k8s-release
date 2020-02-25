package integration_tests_test

import (
	local_k8s_cluster "code.cloudfoundry.org/local-k8s-cluster"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"
)

var _ = Describe("Integration", func() {
	var expectedFileContents = "hi"
	password := "pass"

	BeforeEach(func() {
		var podIP string
		username := "user"
		share := "share"

		By("deploying a smb server", func() {
			command := fmt.Sprintf(`[ "/sbin/tini", "--", "/usr/bin/samba.sh", "-p","-u","%s;%s","-s","%s;/export;no;no;no;%s","-p","-S" ]`, username, password, share, username)
			overrides := fmt.Sprintf(`{"spec": {"template":  {"spec": {"containers": [{"name": "test-smb1", "command": %s, "image": "dperson/samba", "securityContext":{"privileged":true}, "ports": [{"containerPort": 139, "protocol": "TCP"}, {"containerPort": 445, "protocol": "TCP"}]}]}}}}`, command)
			local_k8s_cluster.Kubectl("run", "--overrides", overrides, "--image", "dperson/samba", "test-smb1")
			Eventually(func() string {
				podIP = local_k8s_cluster.Kubectl("get", "pods", "-l", "run=test-smb1", "-o", "jsonpath={.items[0].status.podIPs[0].ip}")
				return podIP
			}, 10 * time.Minute, 2 * time.Second).Should(Not(Equal("")))
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
			share := fmt.Sprintf("//%s/%s", podIP, share)
			requestPayload := fmt.Sprintf(`{ "service_id": "123", "plan_id": "plan-id", "parameters": {"share": "%s", "username": "%s", "password": "%s"} }`, share, username, password)
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

		templ, err := template.New("writer_pod.yaml").ParseFiles("./assets/writer_pod.yaml")
		Expect(err).NotTo(HaveOccurred())
		interpolatedWritedPodYaml, err := ioutil.TempFile(os.TempDir(), "writer_pod_yaml")
		Expect(err).NotTo(HaveOccurred())

		err = templ.Execute(interpolatedWritedPodYaml, struct {
			ExpectedFileContents string
			InstanceID           string
		}{
			expectedFileContents, instanceID,
		})
		Expect(err).NotTo(HaveOccurred())

		interpolatedYamlPath := interpolatedWritedPodYaml.Name()
		Expect(local_k8s_cluster.Kubectl("apply", "-f", interpolatedYamlPath)).To(ContainSubstring("created"))
		Expect(local_k8s_cluster.Kubectl("apply", "-f", "./assets/reader_pod.yaml")).To(ContainSubstring("created"))

		Eventually(func() string {
			return local_k8s_cluster.Kubectl("get", "events", "-n", "cf-workloads")
		}, 10 * time.Minute, 2 * time.Second).Should(ContainSubstring("Started container integration-test-reader"))

		mountCommand := fmt.Sprintf("mkdir /instance1 && mount -t cifs -o username=%s,password=%s //%s/%s /instance1", username, password, podIP, share)
		local_k8s_cluster.Kubectl("exec", "-n", "cf-workloads", "-i", "integration-test-reader", "--", "bash", "-c", mountCommand)
	})

	It("mounts the share to the pod", func() {
		By("the file contents written by a pod with a pvc (created by the broker) should be written to the smb share", func(){
			Eventually(func() string {
				return local_k8s_cluster.Kubectl("exec", "-n", "cf-workloads", "-i", "integration-test-reader", "--", "bash", "-c", "cat /instance1/foo || true")
			}, 10 * time.Minute, 2 * time.Second).Should(ContainSubstring(expectedFileContents))
		})

		By("logs, and sanitizes secrets", func(){
			Expect(local_k8s_cluster.Kubectl("logs", "-l", "app=csi-nodeplugin-smbplugin", "-c", "smb")).To(ContainSubstring("***stripped***"))
		})

		By("never logs the password", func(){
			Expect(local_k8s_cluster.Kubectl("logs", "-l", "app=csi-nodeplugin-smbplugin", "-c", "smb")).NotTo(ContainSubstring(password))
			Expect(local_k8s_cluster.Kubectl("logs", "-l", "app=smb-broker")).NotTo(ContainSubstring(password))
		})
	})
})
