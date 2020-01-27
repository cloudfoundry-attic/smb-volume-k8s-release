package main

import (
	"context"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	return &csi.NodePublishVolumeResponse{}, nil
}

func (noOpNodeServer) NodeUnpublishVolume(c context.Context, r *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	if r.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf(errorFmt, "TargetPath"))
	}

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