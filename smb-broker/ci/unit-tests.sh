#!/bin/bash
set -ex

export PATH=~/go/bin:$PATH

mkdir -p /etc/docker
cat <<EOF >/etc/docker/daemon.json
{
  "storage-driver": "aufs"
}
EOF

/usr/local/bin/start-docker &

rc=1
for i in $(seq 1 100); do
  echo waiting for docker to come up...
  set +e
  docker info
  rc=$?
  set -e
  if [ "$rc" -eq "0" ]; then
      break
  fi
  sleep 1
done

if [ "$rc" -ne "0" ]; then
  exit 1
fi

make --directory=smb-volume-k8s-release/smb-broker test
pkill dockerd