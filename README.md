# smb-volume-k8s-release
This is a repository in development, pre alpha, not feature complete.
This contains helm releases that packages 

- an smb broker
- a smb csi driver
- eirini persi

# Deploying to Cloud Foundry
## Prerequisites
- Install Cloud Foundry on a k8s cluster

## Installing eirini-persi (eirini-ext volume services extension)
```
kubectl create namespace eirini
cd eirini-persi
make build
```

## Installing smb-csi-driver
```
cd smb-csi-driver
make kustomize
```

## Installing smbbroker
```
cd smb-broker
make helm
broker_ip=`kubectl get svc -n default smb-broker -o jsonpath='{.spec.clusterIP}'`
cf create-service-broker smbbroker admin admin http://${broker_ip}
cf enable-service-access smb
```

## Push pora and bind and smb volume
```
cd /tmp
cf push pora -o cfpersi/pora --no-start
cf create-service smb Existing mysmb -c '{"share": "//persi.file.core.windows.net/testshare", "username": "persi", "password": "<password>" }'
cf bind-service pora mysmb
cf start pora
curl https://pora.<system-domain>/write
```

## Testing broker
```
cd smb-broker
make test
```

## Testing csi driver
```
cd smb-csi-driver
make fly
```
