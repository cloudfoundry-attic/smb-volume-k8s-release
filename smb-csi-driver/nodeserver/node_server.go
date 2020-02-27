package nodeserver

import (
	"code.cloudfoundry.org/goshims/execshim"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/lager"
	"context"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"os"
	"sync"
)

var errorFmt = "Error: a required property [%s] was not provided"
var defaultMountOptions = "uid=2000,gid=2000"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o ../smb-csi-driverfakes/fake_csi_driver_store.go . CSIDriverStore
type CSIDriverStore interface {
	Create(string, error)
	Delete(string)
	Get(string) (error, bool)
}

func NewStore() CSIDriverStore {
	return &CheckParallelCSIDriverRequests{store: map[string]error{}}
}

type CheckParallelCSIDriverRequests struct {
	store map[string]error
}

func (c *CheckParallelCSIDriverRequests) Get(k string) (error, bool) {
	val, ok := c.store[k]
	return val, ok
}

func (c *CheckParallelCSIDriverRequests) Create(k string, v error) {
	c.store[k] = v
}

func (c *CheckParallelCSIDriverRequests) Delete(k string) {
	delete(c.store, k)
}

type smbNodeServer struct {
	logger         lager.Logger
	execshim       execshim.Exec
	osshim         osshim.Os
	csiDriverStore CSIDriverStore
	lock           *sync.Mutex
}

func NewNodeServer(logger lager.Logger, execshim execshim.Exec, osshim osshim.Os, csiDriverStore CSIDriverStore) csi.NodeServer {
	return &smbNodeServer{
		logger, execshim, osshim, csiDriverStore, &sync.Mutex{},
	}
}

func (smbNodeServer) NodeGetCapabilities(context.Context, *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return &csi.NodeGetCapabilitiesResponse{}, nil
}

func (smbNodeServer) NodeStageVolume(context.Context, *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	panic("implement me")
}

func (smbNodeServer) NodeUnstageVolume(context.Context, *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	panic("implement me")
}

func (n smbNodeServer) NodePublishVolume(c context.Context, r *csi.NodePublishVolumeRequest) (_ *csi.NodePublishVolumeResponse, err error) {
	n.lock.Lock()
	defer func() {
		n.lock.Unlock()
	}()

	res, found := n.csiDriverStore.Get(r.TargetPath)
	if found {
		if res == nil {
			return &csi.NodePublishVolumeResponse{}, nil
		}
		return nil, res
	}

	defer func() {
		n.csiDriverStore.Create(r.TargetPath, err)
	}()

	if r.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf(errorFmt, "VolumeCapability"))
	}

	err = os.MkdirAll(r.TargetPath, os.ModePerm)
	if err != nil {
		n.logger.Error("create-targetpath-fail", err)
	}

	share := r.GetVolumeContext()["share"]
	username := r.GetSecrets()["username"]
	password := r.GetSecrets()["password"]

	mountOptions := fmt.Sprintf("%s,username=%s,password=%s", defaultMountOptions, username, password)

	n.logger.Info("started mount", lager.Data{"share": share})
	cmdshim := n.execshim.Command("mount", "-t", "cifs", "-o", mountOptions, share, r.TargetPath)
	combinedOutput, err := cmdshim.CombinedOutput()
	if err != nil {
		n.logger.Error("mount-failed", err, lager.Data{"combinedOutput": string(combinedOutput)})
		return nil, status.Error(codes.Internal, err.Error())
	}
	n.logger.Info("finished mount", lager.Data{"share": share})

	return &csi.NodePublishVolumeResponse{}, nil
}

func (n smbNodeServer) NodeUnpublishVolume(c context.Context, r *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	if r.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf(errorFmt, "TargetPath"))
	}

	n.logger.Info("about to remove dir")

	cmdshim := n.execshim.Command("umount", r.TargetPath)
	err := cmdshim.Start()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	n.logger.Info("started umount")

	err = cmdshim.Wait()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	n.logger.Info("finished umount")

	err = n.osshim.Remove(r.TargetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	n.logger.Info("removed dir")

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (smbNodeServer) NodeGetVolumeStats(context.Context, *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	panic("implement me")
}

func (smbNodeServer) NodeExpandVolume(context.Context, *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	panic("implement me")
}

func (s smbNodeServer) NodeGetInfo(context.Context, *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	nodeId, err := s.osshim.Hostname()
	if err != nil {
		return nil, err
	}

	return &csi.NodeGetInfoResponse{
		NodeId: nodeId,
	}, nil
}
