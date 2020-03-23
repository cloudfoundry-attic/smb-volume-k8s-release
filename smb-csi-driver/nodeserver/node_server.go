package nodeserver

import (
	"code.cloudfoundry.org/goshims/execshim"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/lager"
	"context"
	"crypto/sha256"
	"encoding/json"
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
	Create(string, *csi.NodePublishVolumeRequest, error)
	Delete(string)
	Get(string, *csi.NodePublishVolumeRequest) (err error, exists bool, optionsMatch bool)
}

func NewStore() CSIDriverStore {
	return &CheckParallelCSIDriverRequests{store: map[string]volumeInfo{}}
}

type volumeInfo struct {
	err error
	hash [32]byte
}

type CheckParallelCSIDriverRequests struct {
	store map[string]volumeInfo
}

func (c *CheckParallelCSIDriverRequests) Get(targetPath string, k *csi.NodePublishVolumeRequest) (err error, exists bool, optionsMatch bool) {
	options, _ := json.Marshal(k.VolumeContext)
	hash := sha256.Sum256(options)

	if val, ok := c.store[targetPath]; ok {
		if val.hash == hash {
			return val.err, ok, true
		}
		return val.err, ok, false
	}
	return nil, false, false
}

func (c *CheckParallelCSIDriverRequests) Create(targetPath string, k *csi.NodePublishVolumeRequest, v error) {
	options, _ := json.Marshal(k.VolumeContext)
	hash := sha256.Sum256(options)
	c.store[targetPath] = volumeInfo{v, hash}
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

	err, found, optionsMatch := n.csiDriverStore.Get(r.TargetPath, r)
	if found {
		if optionsMatch == false {
			return &csi.NodePublishVolumeResponse{}, status.Error(codes.AlreadyExists, "options mismatch")
		}

		if err == nil {
			return &csi.NodePublishVolumeResponse{}, nil
		}
		return nil, err
	}

	defer func() {
		n.csiDriverStore.Create(r.TargetPath, r, err)
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

func (n smbNodeServer) NodeUnpublishVolume(c context.Context, r *csi.NodeUnpublishVolumeRequest) (_ *csi.NodeUnpublishVolumeResponse, err error) {
	n.lock.Lock()
	defer func() {
		n.lock.Unlock()
	}()

	defer func() {
		n.csiDriverStore.Delete(r.TargetPath)
	}()

	if r.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf(errorFmt, "TargetPath"))
	}

	n.logger.Info("about to remove dir")

	cmdshim := n.execshim.Command("umount", r.TargetPath)
	err = cmdshim.Start()
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
