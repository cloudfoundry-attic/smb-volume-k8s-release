package nodeserver_test

import (
	"code.cloudfoundry.org/goshims/execshim/exec_fake"
	. "code.cloudfoundry.org/smb-csi-driver/nodeserver"
	"context"
	"errors"
	"github.com/container-storage-interface/spec/lib/go/csi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NodeServer", func() {

	var (
		nodeServer csi.NodeServer

		fakeExec *exec_fake.FakeExec
		fakeCmd *exec_fake.FakeCmd
	)

	BeforeEach(func() {
		fakeExec = &exec_fake.FakeExec{}
		fakeCmd = &exec_fake.FakeCmd{}
		fakeExec.CommandReturns(fakeCmd)

		nodeServer = NewNodeServer(fakeExec)
	})

	Describe("#NodePublishVolume", func() {

		var (
			ctx context.Context
			request *csi.NodePublishVolumeRequest
			err error
		)

		BeforeEach(func() {
			request = &csi.NodePublishVolumeRequest{
				VolumeCapability : &csi.VolumeCapability{},
				TargetPath: "/tmp/target_path",
				VolumeContext: map[string]string{
					"share": "//server/export",
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

			It("should return a error", func(){
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("rpc error: code = InvalidArgument desc = Error: a required property [VolumeCapability] was not provided"))
			})
		})

		Context("when making the target directory already exists", func() {
			BeforeEach(func(){
				request.TargetPath = "/tmp"
			})

			It("should report a warning", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		PContext("when making the target directory already that already exists with different arguments", func() {
			BeforeEach(func(){
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
				Expect(args).To(ContainElements("-t", "cifs", "-o", "username=user1,password=pass1", "//server/export", request.TargetPath))
				Expect(fakeCmd.StartCallCount()).To(Equal(1))
				Expect(fakeCmd.WaitCallCount()).To(Equal(1))
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
	})

	Describe("#NodeUnpublishVolume", func() {
		var (
			ctx     context.Context
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
		})
	})
})
