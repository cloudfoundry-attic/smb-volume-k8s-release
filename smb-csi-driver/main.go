package main

import (
	"context"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
)

var errorFmt = "Error: a required property [%s] was not provided"

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 2910))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	csi.RegisterIdentityServer(grpcServer, &noOpIdentityServer{})
	csi.RegisterNodeServer(grpcServer, &noOpNodeServer{})

	err = grpcServer.Serve(lis)
	if err != nil {
		panic(err)
	}
}

type noOpIdentityServer struct {
}

func (*noOpIdentityServer) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	return &csi.GetPluginInfoResponse{
		Name: "A0A",
	}, nil
}
func (*noOpIdentityServer) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	return &csi.GetPluginCapabilitiesResponse{}, nil
}

func (*noOpIdentityServer) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	return &csi.ProbeResponse{}, nil
}

type noOpNodeServer struct {}

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
