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
1. Edit the example (in `./example/pv.yaml`) to use your SMB server:
- `//SERVER/SHARE`: the SMB address of your server and share
- `USERNAME`: username for the share
- `PASSWORD`: password for the share
- `mountOptions`: (optional) supported mount options are uid, gid and vers 

1. Deploy the example
```bash
kubectl apply -f ./example/pv.yaml
```
1. Use the example nginx app by writing a file to the mounted directory
```bash
kubectl exec -it nginx bash
> echo hello > /usr/share/nginx/html/index.html
> apt-get update && apt-get install -y curl
> curl localhost:80
> hello
```

# Testing
```
cd smb-csi-driver
make fly
```