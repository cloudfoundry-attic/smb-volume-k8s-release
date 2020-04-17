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

function kill_docker() {
    pkill dockerd
}
trap kill_docker EXIT

pushd cf-volume-services-acceptance-tests/assets/pora
  pack build "pora" --builder cloudfoundry/cnb:bionic  --buildpack paketo-buildpacks/go-compiler
  pack inspect-image pora --bom | jq -r '.local[0].version' > /tmp/go-version
  cat /tmp/go-version
popd

cp /tmp/go-version go-version/