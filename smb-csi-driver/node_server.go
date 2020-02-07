package main

import (
	"context"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"os"
	"os/exec"
)

var errorFmt = "Error: a required property [%s] was not provided"

type noOpNodeServer struct{}

func (noOpNodeServer) NodeGetCapabilities(context.Context, *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return &csi.NodeGetCapabilitiesResponse{}, nil
}

func (noOpNodeServer) NodeStageVolume(c context.Context, r *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return &csi.NodeStageVolumeResponse{}, nil
}

func (noOpNodeServer) NodeUnstageVolume(context.Context, *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	panic("implement me")
}

func (noOpNodeServer) NodePublishVolume(c context.Context, r *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	if r.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf(errorFmt, "VolumeCapability"))
	}

	err := os.MkdirAll(r.TargetPath, os.ModePerm)
	if err != nil {
		println(err.Error())
	}
	serverIP := r.GetVolumeContext()["server"]
	share := r.GetVolumeContext()["share"]
	username := r.GetVolumeContext()["username"]
	password := r.GetVolumeContext()["password"]

	fmt.Println(fmt.Sprintf("target path: %s", r.TargetPath))

	uncPath := fmt.Sprintf("//%s%s", serverIP, share)
	fmt.Println(fmt.Sprintf("about to mount to %s", uncPath))

	mountOptions := fmt.Sprintf("username=%s,password=%s", username, password)

	cmd := exec.Command("mount", "-t", "cifs", "-o", mountOptions, uncPath, r.TargetPath)
	err = cmd.Start()
	if err != nil {
		println(err.Error())
	}
	fmt.Println(fmt.Sprintf("started mount to %s", uncPath))

	err = cmd.Wait()
	if err != nil {
		println(err.Error())
	}
	fmt.Println(fmt.Sprintf("finished mount to %s", uncPath))

	return &csi.NodePublishVolumeResponse{}, nil
}

func (noOpNodeServer) NodeUnpublishVolume(c context.Context, r *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	if r.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf(errorFmt, "TargetPath"))
	}

	fmt.Println(fmt.Sprintf("about to remove dir"))

	cmd := exec.Command("umount", r.TargetPath)
	err := cmd.Start()
	if err != nil {
		println(err.Error())
	}
	fmt.Println("started umount")

	err = cmd.Wait()
	if err != nil {
		println(err.Error())
	}
	fmt.Println("finished umount")

	err = os.Remove(r.TargetPath)
	if err != nil {
		println(err.Error())
	}

	fmt.Println(fmt.Sprintf("removed dir"))

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (noOpNodeServer) NodeGetVolumeStats(context.Context, *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	panic("implement me")
}

func (noOpNodeServer) NodeExpandVolume(context.Context, *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	panic("implement me")
}

func (noOpNodeServer) NodeGetInfo(context.Context, *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return &csi.NodeGetInfoResponse{
		NodeId: "node-id",
	}, nil
}