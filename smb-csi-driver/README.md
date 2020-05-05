# Installation
## Installing alpha version
```
kubectl apply --kustomize "https://github.com/cloudfoundry/smb-volume-k8s-release/smb-csi-driver/deploy/overlays/alpha/?ref=master"
```

## Installing latest dev version
```
kubectl apply --kustomize "https://github.com/cloudfoundry/smb-volume-k8s-release/smb-csi-driver/deploy/overlays/dev/?ref=master"
```

# Usage example

1. Deploy the test samba server
``` 
kubectl apply -f ./example/samba.yml
```
1. Deploy a sample pv, pvc and pod
```bash
kubectl apply -f ./example/pv.yaml
```
1. Use the sample nginx app by writing a file to the mounted directory
```bash
kubectl exec -it nginx bash
> echo hello > /usr/share/nginx/html/index.html
> apt-get update && apt-get install -y curl
> curl localhost:80
> hello
```

1. Use with your own SMB server by editing the `pv.yaml` fields as follows:
- `share`: the SMB address of your server and share
- `username`: username for the share
- `password`: password for the share
- `mountOptions`: (optional) supported mount options are uid, gid and vers 

# Testing
```
cd smb-csi-driver
make fly
```