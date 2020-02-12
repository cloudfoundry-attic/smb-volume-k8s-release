package nodeserver

import (
	"code.cloudfoundry.org/goshims/execshim"
	"context"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"os"
)

var errorFmt = "Error: a required property [%s] was not provided"

type noOpNodeServer struct {
	execshim execshim.Exec
}

func NewNodeServer(execshim execshim.Exec) csi.NodeServer {
	return &noOpNodeServer{
		execshim,
	}
}

func (noOpNodeServer) NodeGetCapabilities(context.Context, *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return &csi.NodeGetCapabilitiesResponse{}, nil
}

func (noOpNodeServer) NodeStageVolume(c context.Context, r *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return &csi.NodeStageVolumeResponse{}, nil
}

func (noOpNodeServer) NodeUnstageVolume(context.Context, *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	panic("implement me")
}

func (n noOpNodeServer) NodePublishVolume(_ context.Context, r *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	if r.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf(errorFmt, "VolumeCapability"))
	}

	err := os.MkdirAll(r.TargetPath, os.ModePerm)
	if err != nil {
		println(err.Error())
	}
	share := r.GetVolumeContext()["share"]
	username := r.GetVolumeContext()["username"]
	password := r.GetVolumeContext()["password"]

	log.Printf("local target path: %s", r.TargetPath)

	mountOptions := fmt.Sprintf("username=%s,password=%s", username, password)

	cmdshim := n.execshim.Command("mount", "-t", "cifs", "-o", mountOptions, share, r.TargetPath)
	err = cmdshim.Start()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	fmt.Println(fmt.Sprintf("started mount to %s", share))

	err = cmdshim.Wait()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	fmt.Println(fmt.Sprintf("finished mount to %s", share))

	return &csi.NodePublishVolumeResponse{}, nil
}

func (n noOpNodeServer) NodeUnpublishVolume(c context.Context, r *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	if r.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf(errorFmt, "TargetPath"))
	}

	log.Printf("about to remove dir")

	cmdshim := n.execshim.Command("umount", r.TargetPath)
	err := cmdshim.Start()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	log.Print("started umount")

	err = cmdshim.Wait()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	log.Printf("finished umount")

	err = os.Remove(r.TargetPath)
	if err != nil {
		println(err.Error())
	}

	log.Printf("removed dir")

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
