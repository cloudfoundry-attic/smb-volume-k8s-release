package identityserver

import (
	"context"
	"github.com/container-storage-interface/spec/lib/go/csi"
)

type smbIdentityServer struct {
}

func (*smbIdentityServer) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	return &csi.GetPluginInfoResponse{
		Name: "org.cloudfoundry.smb",
	}, nil
}
func (*smbIdentityServer) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	return &csi.GetPluginCapabilitiesResponse{}, nil
}

func (*smbIdentityServer) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	return &csi.ProbeResponse{}, nil
}

