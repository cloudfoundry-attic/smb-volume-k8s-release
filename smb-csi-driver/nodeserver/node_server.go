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
	"strings"
	"sync"
)

var errorFmt = "Error: a required property [%s] was not provided"
var defaultMountOptions = "uid=1000,gid=1000"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o ../smb-csi-driverfakes/fake_csi_driver_store.go . CSIDriverStore
type CSIDriverStore interface {
	Create(string, *csi.NodePublishVolumeRequest) error
	Delete(string)
	Get(string, *csi.NodePublishVolumeRequest) (exists bool, optionsMatch bool, err error)
}

func NewStore() CSIDriverStore {
	return &CheckParallelCSIDriverRequests{store: map[string]volumeInfo{}}
}

type volumeInfo struct {
	hash [32]byte
}

type CheckParallelCSIDriverRequests struct {
	store map[string]volumeInfo
}

func (c *CheckParallelCSIDriverRequests) Get(targetPath string, k *csi.NodePublishVolumeRequest) (exists bool, optionsMatch bool, err error) {
	options, err := json.Marshal(k.VolumeContext)
	if err != nil {
		return true, true, err
	}
	hash := sha256.Sum256(options)

	if val, ok := c.store[targetPath]; ok {
		if val.hash == hash {
			return ok, true, nil
		}
		return ok, false, nil
	}
	return false, false, nil
}

func (c *CheckParallelCSIDriverRequests) Create(targetPath string, k *csi.NodePublishVolumeRequest) error {
	options, err := json.Marshal(k.VolumeContext)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(options)
	c.store[targetPath] = volumeInfo{ hash}
	return nil
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

func (n smbNodeServer) NodePublishVolume(c context.Context, r *csi.NodePublishVolumeRequest) (_ *csi.NodePublishVolumeResponse, opErr error) {
	n.lock.Lock()
	defer func() {
		n.lock.Unlock()
	}()

	found, optionsMatch, err := n.csiDriverStore.Get(r.TargetPath, r)
	if err != nil {
		return &csi.NodePublishVolumeResponse{}, err
	}
	if found {
		if optionsMatch == false {
			return &csi.NodePublishVolumeResponse{}, status.Error(codes.AlreadyExists, "options mismatch")
		}

		if opErr == nil {
			return &csi.NodePublishVolumeResponse{}, nil
		}
	}

	defer func() {
		if opErr == nil {
			createErr := n.csiDriverStore.Create(r.TargetPath, r)
			if createErr != nil {
				opErr = createErr
			}
		}
	}()

	if r.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf(errorFmt, "VolumeCapability"))
	}

	opErr = os.MkdirAll(r.TargetPath, os.ModePerm)
	if opErr != nil {
		n.logger.Error("create-targetpath-fail", opErr)
	}

	share := r.GetVolumeContext()["share"]
	username := r.GetSecrets()["username"]
	password := r.GetSecrets()["password"]

	mountOptions := fmt.Sprintf("%s,username=%s,password=%s", defaultMountOptions, username, password)

	vers, ok := r.GetVolumeContext()["vers"]
	if ok {
		if strings.Contains(vers, ",") {
			return nil, status.Error(codes.InvalidArgument, "Error: invalid VolumeContext value for 'vers'")
		}
		mountOptions += ",vers=" + vers
	}

	n.logger.Info("started mount", lager.Data{"share": share})
	cmdshim := n.execshim.Command("mount", "-t", "cifs", "-o", mountOptions, share, r.TargetPath)
	combinedOutput, opErr := cmdshim.CombinedOutput()
	if opErr != nil {
		n.logger.Error("mount-failed", opErr, lager.Data{"combinedOutput": string(combinedOutput)})
		return nil, status.Error(codes.Internal, opErr.Error())
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
