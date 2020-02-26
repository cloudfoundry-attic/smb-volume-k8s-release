package local_k8s_cluster

const SPIN_UP_LOCAL_REGISTRY_BASH = `#!/bin/bash
set -x

reg_name='kind-registry'
reg_port='5000'
ip_fmt='{{.NetworkSettings.IPAddress}}'

# add host mapping into node container
docker exec "$1" sh -c "echo $(docker inspect -f "${ip_fmt}" "${reg_name}") registry >> /etc/hosts"`
