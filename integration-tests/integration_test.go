package integration_tests_test

import (
	local_k8s_cluster "code.cloudfoundry.org/smb-volume-k8s-local-cluster"
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
		username := "user"
		share := "//smb1.default/share"

		By("deploying a smb server", func() {
			local_k8s_cluster.Kubectl("apply", "-f", "./assets/samba.yml")
			Eventually(func() string {
				return local_k8s_cluster.Kubectl("get", "po", "-l=app.kubernetes.io/name=test-smb1")
			}, 10 * time.Minute, 2 * time.Second).Should((ContainSubstring("Running")))

			Eventually(func() string {
				return local_k8s_cluster.Kubectl("logs", "-l=app.kubernetes.io/name=test-smb1")
			}, 10 * time.Minute, 2 * time.Second).Should((ContainSubstring("finished starting up")))
		})

		var resp *http.Response

		Eventually(func() string {
			resp, _ = http.DefaultClient.Get("http://foo:bar@localhost/v2/catalog")
			if resp == nil {
				return ""
			}
			return resp.Status
		}).Should(Equal("200 OK"))

		instanceID := "instance1"
		bindingID := "binding1"
		serviceID := "6cb45412-8161-44ec-b462-e3fd08f55448"
		planID := "e805eb41-4fb4-485a-9066-b0edf57b90b3"

		Eventually(func() string {
			requestPayload := fmt.Sprintf(`{ "service_id": "%s", "plan_id": "%s", "parameters": {"share": "%s", "username": "%s", "password": "%s"} }`, serviceID, planID, share, username, password)
			request, err := http.NewRequest("PUT", fmt.Sprintf("http://foo:bar@localhost/v2/service_instances/%s", instanceID), strings.NewReader(requestPayload))
			Expect(err).NotTo(HaveOccurred())

			resp, _ = http.DefaultClient.Do(request)
			if resp == nil {
				return ""
			}
			return resp.Status
		}).Should(Equal("201 Created"))

		Eventually(func() string {
			request, err := http.NewRequest("PUT", fmt.Sprintf("http://foo:bar@localhost/v2/service_instances/%s/service_bindings/%s", instanceID, bindingID),
				strings.NewReader(fmt.Sprintf(`{"service_id": "%s", "plan_id": "%s", "bind_resource": {"app_guid": "123"}}`, serviceID, planID)))

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

		mountCommand := fmt.Sprintf("mkdir /instance1 && mount -t cifs -o username=%s,password=%s %s /instance1", username, password, share)
		local_k8s_cluster.Kubectl("exec", "-n", "cf-workloads", "-i", "integration-test-reader", "--", "bash", "-c", mountCommand)
	})

	It("mounts the share to the pod", func() {
		By("the file contents written by a pod with a pvc (created by the broker) should be written to the smb share", func(){
			Eventually(func() string {
				return local_k8s_cluster.Kubectl("exec", "-n", "cf-workloads", "-i", "integration-test-reader", "--", "bash", "-c", "cat /instance1/foo || true")
			}, 10 * time.Minute, 2 * time.Second).Should(ContainSubstring(expectedFileContents))
		})

		By("logs, and sanitizes secrets", func(){
			Expect(local_k8s_cluster.Kubectl("logs", "-l", "app=csi-nodeplugin-smbplugin", "-c", "smb", "-n", "cf-smb")).To(ContainSubstring("***stripped***"))
		})

		By("never logs the password", func(){
			Expect(local_k8s_cluster.Kubectl("logs", "-l", "app=csi-nodeplugin-smbplugin", "-c", "smb", "-n", "cf-smb")).NotTo(ContainSubstring(password))
			Expect(local_k8s_cluster.Kubectl("logs", "-l", "app=smb-broker", "-n", "cf-smb")).NotTo(ContainSubstring(password))
		})
	})
})
