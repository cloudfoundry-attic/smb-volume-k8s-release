package main_test

import (
	local_k8s_cluster "code.cloudfoundry.org/smb-volume-k8s-local-cluster"
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"time"
)

type CSIDriversInfo struct {
	APIVersion string   `json:"apiVersion"`
	Items      []Items  `json:"items"`
	Kind       string   `json:"kind"`
	Metadata   Metadata `json:"metadata"`
}
type Annotations struct {
	KubectlKubernetesIoLastAppliedConfiguration string `json:"kubectl.kubernetes.io/last-applied-configuration"`
}
type Items struct {
	APIVersion string        `json:"apiVersion"`
	Kind       string        `json:"kind"`
	Metadata   CSIDriverMeta `json:"metadata"`
	Spec       CSIDriverSpec `json:"spec"`
}
type CSIDriverMeta struct {
	Annotations       Annotations `json:"annotations"`
	CreationTimestamp time.Time   `json:"creationTimestamp"`
	Name              string      `json:"name"`
	ResourceVersion   string      `json:"resourceVersion"`
	SelfLink          string      `json:"selfLink"`
	UID               string      `json:"uid"`
}
type CSIDriverSpec struct {
	AttachRequired       bool     `json:"attachRequired"`
	PodInfoOnMount       bool     `json:"podInfoOnMount"`
	VolumeLifecycleModes []string `json:"volumeLifecycleModes"`
}
type Metadata struct {
	ResourceVersion string `json:"resourceVersion"`
	SelfLink        string `json:"selfLink"`
}
var _ = Describe("CSIDriver object", func() {

	It("should register a CSIDriver object", func() {
		getCsiDriverOutout := local_k8s_cluster.Kubectl("get", "csidriver", "-o", "json")
		csiDriverInfos := CSIDriversInfo{}
		Expect(json.Unmarshal([]byte(getCsiDriverOutout), &csiDriverInfos)).NotTo(HaveOccurred())

		var csiDriver Items
		for _, csiDriverInfo := range csiDriverInfos.Items {
			if csiDriverInfo.Metadata.Name == "org.cloudfoundry.smb" {
				csiDriver = csiDriverInfo
			}
		}

		Expect(csiDriver).NotTo(BeNil())
		Expect(csiDriver.Spec.AttachRequired).To(BeFalse())
		Expect(csiDriver.Spec.PodInfoOnMount).To(BeFalse())
		k8sVersion := os.Getenv("K8S_IMAGE")
		if k8sVersion == "kindest/node:v1.15.7" {
			Expect(csiDriver.Spec.VolumeLifecycleModes).To(BeNil())
		} else {
			Expect(csiDriver.Spec.VolumeLifecycleModes).To(ContainElement("Persistent"))
			Expect(csiDriver.Spec.VolumeLifecycleModes).NotTo(ContainElement("Ephemeral"))
		}
	})
})
