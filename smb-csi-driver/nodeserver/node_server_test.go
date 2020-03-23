package nodeserver_test

import (
	"code.cloudfoundry.org/goshims/execshim/exec_fake"
	"code.cloudfoundry.org/goshims/osshim/os_fake"
	"code.cloudfoundry.org/lager/lagertest"
	. "code.cloudfoundry.org/smb-csi-driver/nodeserver"
	smbcsidriverfakes "code.cloudfoundry.org/smb-csi-driver/smb-csi-driverfakes"
	"context"
	"errors"
	"github.com/container-storage-interface/spec/lib/go/csi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"sync"
)

var _ = Describe("NodeServer", func() {

	var (
		logger     *lagertest.TestLogger
		nodeServer csi.NodeServer
		ctx        context.Context

		fakeOs             *os_fake.FakeOs
		fakeExec           *exec_fake.FakeExec
		fakeCmd            *exec_fake.FakeCmd
		fakeCSIDriverStore *smbcsidriverfakes.FakeCSIDriverStore
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("node-server-test")
		fakeOs = &os_fake.FakeOs{}
		fakeExec = &exec_fake.FakeExec{}
		fakeCmd = &exec_fake.FakeCmd{}
		fakeExec.CommandReturns(fakeCmd)
		fakeCSIDriverStore = &smbcsidriverfakes.FakeCSIDriverStore{}
		ctx = context.Background()

		nodeServer = NewNodeServer(logger, fakeExec, fakeOs, fakeCSIDriverStore)
	})

	Describe("parallel identical #NodePublish requests", func() {
		var (
			request                   *csi.NodePublishVolumeRequest
		)

		BeforeEach(func() {
			request = &csi.NodePublishVolumeRequest{
				VolumeCapability: &csi.VolumeCapability{},
				TargetPath:       "/tmp/target_path",
				VolumeContext: map[string]string{
					"share": "//server/export",
				},
				Secrets: map[string]string{
					"username": "user1",
					"password": "pass1",
				},
			}
			fakeCSIDriverStore.GetReturns(nil, true, true)
			fakeCSIDriverStore.GetReturnsOnCall(0, nil, false, true)


		})

		It("should handle concurrent requests correctly", func(){
			var wg sync.WaitGroup

			wg.Add(10)
			for i:=0; i<10;i++ {
				go func() {
					defer GinkgoRecover()
					defer wg.Done()
					_, err := nodeServer.NodePublishVolume(ctx, request)
					Expect(err).NotTo(HaveOccurred())
				}()
			}
			wg.Wait()

			Expect(fakeCSIDriverStore.CreateCallCount()).To(Equal(1))
			p, k, v := fakeCSIDriverStore.CreateArgsForCall(0)
			Expect(p).To(Equal(request.TargetPath))
			Expect(k).To(Equal(request))
			Expect(v).NotTo(HaveOccurred())
		})
	})

	Describe("#NodePublishVolume", func() {

		var (
			request                   *csi.NodePublishVolumeRequest
			err                       error
			nodePublishVolumeResponse *csi.NodePublishVolumeResponse
		)

		BeforeEach(func() {
			request = &csi.NodePublishVolumeRequest{
				VolumeCapability: &csi.VolumeCapability{},
				TargetPath:       "/tmp/target_path",
				VolumeContext: map[string]string{
					"share": "//server/export",
				},
				Secrets: map[string]string{
					"username": "user1",
					"password": "pass1",
				},
			}
		})

		JustBeforeEach(func() {
			nodePublishVolumeResponse, err = nodeServer.NodePublishVolume(ctx, request)
		})

		Context("when VolumeCapability is not supplied", func() {
			BeforeEach(func() {
				request.VolumeCapability = nil
			})

			It("should return a error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("rpc error: code = InvalidArgument desc = Error: a required property [VolumeCapability] was not provided"))

				Expect(fakeCSIDriverStore.CreateCallCount()).NotTo(BeZero())
				p, k, v := fakeCSIDriverStore.CreateArgsForCall(0)
				Expect(p).To(Equal(request.TargetPath))
				Expect(k).To(Equal(request))
				Expect(v).To(HaveOccurred())
				Expect(v).To(MatchError("rpc error: code = InvalidArgument desc = Error: a required property [VolumeCapability] was not provided"))
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

		Context("given a server, a share, a username and password", func() {

			It("should audit the operation in a map", func() {
				Expect(fakeCSIDriverStore.CreateCallCount()).To(Equal(1))
				p, k, v := fakeCSIDriverStore.CreateArgsForCall(0)

				Expect(p).To(Equal(request.TargetPath))
				Expect(k).To(Equal(request))
				Expect(v).To(BeNil())
			})

			Context("when a second identical request is made", func() {
				BeforeEach(func() {
					fakeCSIDriverStore.GetReturns(nil, true, false)
				})

				It("return the response of the previous request", func() {
					Expect(fakeCSIDriverStore.CreateCallCount()).To(Equal(0))
					Expect(nodePublishVolumeResponse).To(Equal(&csi.NodePublishVolumeResponse{}))
				})
			})

			Context("when a second identical request is made after an error", func() {
				BeforeEach(func() {
					fakeCSIDriverStore.GetReturns(errors.New("I <3 you."), true, true)
				})

				It("return the response of the previous request", func() {
					Expect(fakeCSIDriverStore.CreateCallCount()).To(Equal(0))
					Expect(nodePublishVolumeResponse).To(BeNil())
					Expect(err).To(HaveOccurred())
				})
			})

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

			It("should store the error in case we get called again", func(){
				Expect(fakeCSIDriverStore.CreateCallCount()).NotTo(BeZero())
				p, k, v := fakeCSIDriverStore.CreateArgsForCall(0)
				Expect(p).To(Equal(request.TargetPath))
				Expect(k).To(Equal(request))
				Expect(v).To(HaveOccurred())
				Expect(v).To(MatchError("rpc error: code = Internal desc = cmd-failed"))

			})
		})

		Context("when creating an entry in the store fails", func() {

			BeforeEach(func() {
				fakeCSIDriverStore.CreateReturns(errors.New("hash failure"))
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("hash failure"))
			})
		})

		Context("when a second NodePublish occurs", func() {
			Context("when it uses the same mount options", func() {

				It("return successfully", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when it uses different mount options", func() {
				BeforeEach(func() {
					fakeCSIDriverStore.GetReturns(nil, true, false)
				})

				It("return ALREADY_EXISTS", func() {
					Expect(err).To(MatchError("rpc error: code = AlreadyExists desc = options mismatch"))
				})
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

		It("should remove the publish volume record from the map", func() {
			Expect(fakeCSIDriverStore.DeleteCallCount()).To(Equal(1))
			k := fakeCSIDriverStore.DeleteArgsForCall(0)

			Expect(k).To(Equal("/tmp/target_path"))
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

			It("should remove the publish volume record from the map", func() {
				Expect(fakeCSIDriverStore.DeleteCallCount()).To(Equal(1))
				k := fakeCSIDriverStore.DeleteArgsForCall(0)

				Expect(k).To(Equal("/tmp/target_path"))
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
