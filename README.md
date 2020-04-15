[![Build Status](https://hush-house.pivotal.io/api/v1/teams/cf-volume-services/pipelines/cf-volume-services-k8s/badge)](https://hush-house.pivotal.io/api/v1/teams/cf-volume-services/pipelines/cf-volume-services-k8s/badge)

# smb-volume-k8s-release
This is a repository in development, pre alpha, not feature complete.
This contains kapp applications that package:

- an smb broker
- a smb csi driver
- eirini persi

# Deploying to Cloud Foundry
## Prerequisites
- [Install Cloud Foundry on a k8s cluster](https://github.com/cloudfoundry/cf-for-k8s/blob/master/docs/deploy.md)

## Installing eirini-persi (eirini-ext volume services extension)
```
cd eirini-persi
make kapp
```

## Installing smb-csi-driver
```
cd smb-csi-driver
make kapp
```

## Installing smbbroker
```
cd smb-broker
make kapp
cf create-service-broker smbbroker foo foo http://smb-broker.cf-smb
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
