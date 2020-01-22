# smb-volume-k8s-release
This is a repository in development, pre alpha, not feature complete.
This contains helm releases that packages

- an smb broker
- a smb csi driver


# Deploying to Cloud Foundry
## Prerequisites
- Install Cloud Foundry on a k8s cluster

## Installing eirini-persi (eirini-ext volume services extension)
```
cd eirini-persi
make build
```

## Installing smb-csi-driver
```
cd smb-csi-driver
make build
```

## Installing smbbroker
```
cd smb-broker
make helm
```

## Testing 
```
cd smb-broker
make image-local-registry test
```
