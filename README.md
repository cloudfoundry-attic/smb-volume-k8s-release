# smb-volume-k8s-release
This is a repository in development, pre alpha, not feature complete.
This contains helm releases that packages 

- an smb broker
- a smb csi driver
- eirini persi

# Deploying to Cloud Foundry
## Prerequisites
- Install Cloud Foundry on a k8s cluster

### CF4K8s
- Turn on the feature flag on capi:
Make the local change here.
https://github.com/cloudfoundry/cf-for-k8s/blob/7dc8b899316b7091affe2bf59947b93234143545/config/_ytt_lib/github.com/cloudfoundry/capi-k8s-release/templates/ccng-config.lib.yml#L53
Then re-install to re-apply the changes:
```
cd $HOME/cf-for-k8s/bin
./install-cf.sh /tmp/cf-values.yml
```

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
make kapp
cf create-service-broker smbbroker foo foo http://smb-broker.default
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
