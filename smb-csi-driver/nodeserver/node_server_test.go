package nodeserver_test

import (
	"code.cloudfoundry.org/goshims/execshim/exec_fake"
	"code.cloudfoundry.org/goshims/osshim/os_fake"
	"code.cloudfoundry.org/lager/lagertest"
	. "code.cloudfoundry.org/smb-csi-driver/nodeserver"
	"context"
	"errors"
	"github.com/container-storage-interface/spec/lib/go/csi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o ../smb-csi-driverfakes/fake_configmap_interface.go  k8s.io/client-go/kubernetes/typed/core/v1.ConfigMapInterface

var _ = Describe("NodeServer", func() {

	var (
		logger 		*lagertest.TestLogger
		nodeServer csi.NodeServer
		ctx        context.Context

		fakeOs   *os_fake.FakeOs
		fakeExec *exec_fake.FakeExec
		fakeCmd  *exec_fake.FakeCmd
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("node-server-test")
		fakeOs = &os_fake.FakeOs{}
		fakeExec = &exec_fake.FakeExec{}
		fakeCmd = &exec_fake.FakeCmd{}
		fakeExec.CommandReturns(fakeCmd)
		ctx = context.Background()

		nodeServer = NewNodeServer(logger, fakeExec, fakeOs)
	})

	Describe("#NodePublishVolume", func() {

		var (
			request *csi.NodePublishVolumeRequest
			err     error
		)

		BeforeEach(func() {
			request = &csi.NodePublishVolumeRequest{
				VolumeCapability: &csi.VolumeCapability{},
				TargetPath:       "/tmp/target_path",
				VolumeContext: map[string]string{
					"share":    "//server/export",
				},
				Secrets: map[string]string{
					"username": "user1",
					"password": "pass1",
				},
			}
		})

		JustBeforeEach(func() {
			_, err = nodeServer.NodePublishVolume(ctx, request)
		})

		Context("when VolumeCapability is not supplied", func() {
			BeforeEach(func() {
				request.VolumeCapability = nil
			})

			It("should return a error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("rpc error: code = InvalidArgument desc = Error: a required property [VolumeCapability] was not provided"))
			})
		})

		Context("when making the target directory already exists", func() {
			BeforeEach(func() {
				request.TargetPath = "/tmp"
			})

			It("should report a warning", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		PContext("when making the target directory already that already exists with different arguments", func() {
			BeforeEach(func() {
				request.TargetPath = "/tmp"
			})

			It("should return a ALREADY_EXISTS error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("given a server, a share, a username and password", func() {

			It("should perform a mount", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeExec.CommandCallCount()).To(Equal(1))
				command, args := fakeExec.CommandArgsForCall(0)
				Expect(command).To(Equal("mount"))
				Expect(args).To(ContainElements("-t", "cifs", "-o", "uid=2000,gid=2000,username=user1,password=pass1", "//server/export", request.TargetPath))
			})
		})

		Context("when the command fails to start", func() {

			BeforeEach(func() {
				fakeCmd.CombinedOutputReturns([]byte("some-stdout"), errors.New("cmd-failed"))
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("rpc error: code = Internal desc = cmd-failed"))
			})

			It("should write the error, stdout and stderr to the logs", func() {
				Eventually(logger.Buffer()).Should(Say("some-stdout"))
				Eventually(logger.Buffer()).Should(Say("cmd-failed"))
			})
		})
	})

	Describe("#NodeUnpublishVolume", func() {
		var (
			request *csi.NodeUnpublishVolumeRequest
			err     error
		)

		JustBeforeEach(func() {
			_, err = nodeServer.NodeUnpublishVolume(ctx, request)
		})

		BeforeEach(func() {
			request = &csi.NodeUnpublishVolumeRequest{
				TargetPath: "/tmp/target_path",
			}
		})

		It("should unpublish the target path", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeExec.CommandCallCount()).To(Equal(1))
			command, args := fakeExec.CommandArgsForCall(0)
			Expect(command).To(Equal("umount"))
			Expect(args).To(ContainElements(request.TargetPath))
			Expect(fakeCmd.StartCallCount()).To(Equal(1))
			Expect(fakeCmd.WaitCallCount()).To(Equal(1))
		})

		Context("when target path is not provided", func() {
			BeforeEach(func() {
				request = &csi.NodeUnpublishVolumeRequest{
					TargetPath: "",
				}
			})

			It("should return a meaningful error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("rpc error: code = InvalidArgument desc = Error: a required property [TargetPath] was not provided"))
			})
		})

		Context("when the command fails to start", func() {

			BeforeEach(func() {
				fakeCmd.StartReturns(errors.New("start-failed"))
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("rpc error: code = Internal desc = start-failed"))
			})
		})

		Context("when the command fails to wait", func() {

			BeforeEach(func() {
				fakeCmd.WaitReturns(errors.New("wait-failed"))
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("rpc error: code = Internal desc = wait-failed"))
			})
		})

		Context("when removing the unmounted target path fails", func() {
			BeforeEach(func() {
				fakeOs.RemoveReturns(errors.New("remove-failed"))
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("rpc error: code = Internal desc = remove-failed"))
			})
		})
	})

	Describe("#NodeGetCapabilities", func() {
		It("should return no capabilities, and no errors", func() {
			resp, err := nodeServer.NodeGetCapabilities(ctx, &csi.NodeGetCapabilitiesRequest{})

			Expect(err).NotTo(HaveOccurred())
			Expect(resp).To(Equal(&csi.NodeGetCapabilitiesResponse{}))
		})
	})

	Describe("#NodeGetInfo", func() {

		var (
			resp *csi.NodeGetInfoResponse
			err  error
		)
		BeforeEach(func() {
			fakeOs.HostnameReturns("hostWithTheMost", nil)
		})

		JustBeforeEach(func() {
			resp, err = nodeServer.NodeGetInfo(ctx, &csi.NodeGetInfoRequest{})
		})

		It("should return the hostname as the node id", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).To(Equal(&csi.NodeGetInfoResponse{NodeId: "hostWithTheMost"}))
			Expect(fakeOs.HostnameCallCount()).To(Equal(1))
		})

		Context("when unable to retrieve the hostname", func() {
			BeforeEach(func() {
				fakeOs.HostnameReturns("", errors.New("catastrophe!"))
			})

			It("should handle OS errors correctly", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("catastrophe"))
			})

		})
	})
})
