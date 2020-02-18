module code.cloudfoundry.org/smb-csi-driver

go 1.13

require (
	code.cloudfoundry.org/goshims v0.1.0
	code.cloudfoundry.org/local-k8s-cluster v0.0.0
	github.com/container-storage-interface/spec v1.2.0
	github.com/kubernetes-csi/csi-test/v3 v3.0.0
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	google.golang.org/grpc v1.27.1
	k8s.io/api v0.0.0
	k8s.io/apimachinery v0.17.0
	k8s.io/kubernetes v1.18.0-alpha.2.0.20200203095321-4c3aa3f26b84
)

replace (
	code.cloudfoundry.org/local-k8s-cluster => ../local-k8s-cluster
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	k8s.io/api => k8s.io/api v0.0.0-20200202064633-3d77e12e1dcd
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20200131034431-1ed3fae1c0f1
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.0-alpha.2.0.20200131032148-f30c02351710
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20200131033534-9dae63f1bed9
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20200202075133-d89ff7d77990
	k8s.io/client-go => k8s.io/client-go v0.0.0-20200202104520-21de178e1daf
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20200202081617-5ca9197f3518
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20200202081248-c3653675c32a
	k8s.io/code-generator => k8s.io/code-generator v0.18.0-alpha.2.0.20200130061103-7dfd5e9157ef
	k8s.io/component-base => k8s.io/component-base v0.0.0-20200131033216-6bcd25baa4f4
	k8s.io/cri-api => k8s.io/cri-api v0.18.0-alpha.2.0.20200130075657-0f4d8f7cdaf8
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20200202081942-47f44efa8d02
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20200203105121-baa892dde80d
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20200131035424-a52a68e0bb54
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20200131035021-daa77faad48f
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20200131035304-a91166310c94
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20200202083509-154f6764baf7
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20200202080222-5356f792ef2c
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20200202082337-b0dccdf763a2
	k8s.io/metrics => k8s.io/metrics v0.0.0-20200202074655-99a7c0b41f45
	k8s.io/node-api => k8s.io/node-api v0.0.0-20200202082757-ca36ebc6532d
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20200131033911-832555428e82
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.0.0-20200202075519-81d99479e6e2
)
