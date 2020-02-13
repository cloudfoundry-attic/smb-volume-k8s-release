package identityserver_test

import (
	. "code.cloudfoundry.org/smb-csi-driver/identityserver"
	"context"
	"github.com/container-storage-interface/spec/lib/go/csi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IdentityServer", func() {
	var (
		ctx    context.Context
		server csi.IdentityServer
	)

	BeforeEach(func() {
		ctx = context.Background()
		server = NewSmbIdentityServer()
	})

	Describe("#GetPluginInfo", func() {
		It("should return the correct plugin id", func() {
			resp, err := server.GetPluginInfo(ctx, &csi.GetPluginInfoRequest{})

			Expect(err).NotTo(HaveOccurred())
			Expect(resp).To(Equal(&csi.GetPluginInfoResponse{Name: "org.cloudfoundry.smb"}))
		})
	})

	Describe("#GetPluginCapabilities", func() {
		It("should return no plugin capability", func() {
			resp, err := server.GetPluginCapabilities(ctx, &csi.GetPluginCapabilitiesRequest{})

			Expect(err).NotTo(HaveOccurred())
			Expect(resp).To(Equal(&csi.GetPluginCapabilitiesResponse{}))
		})
	})

	Describe("#Probe", func() {
		It("should return a 'normal' status", func() {
			resp, err := server.Probe(ctx, &csi.ProbeRequest{})

			Expect(err).NotTo(HaveOccurred())
			Expect(resp).To(Equal(&csi.ProbeResponse{}))
		})
	})
})
