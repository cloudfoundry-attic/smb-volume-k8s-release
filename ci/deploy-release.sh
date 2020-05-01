#!/bin/bash -eux

#!/bin/bash -eu

ZONE="us-west1-a"
echo ${GCP_KEY} > /tmp/gcp.json

gcloud auth activate-service-account  --key-file /tmp/gcp.json
gcloud config set project cff-diego-persistence

CLUSTER_NAME=$(cat /dev/urandom | tr -dc 'a-z' | fold -w 32 | head --bytes 10)
if [ "$CLUSTER_NAME" == "" ]; then
  CLUSTER_NAME=default-ci
fi

gcloud -q container clusters create ${CLUSTER_NAME} --zone ${ZONE} --machine-type n1-standard-4 --image-type=ubuntu --num-nodes=3

CLUSTER_IP_NAME=${CLUSTER_NAME}
CLUSTER_DNS=${CLUSTER_NAME}.persi.cf-app.com

gcloud container clusters get-credentials ${CLUSTER_NAME} --zone ${ZONE} --project cff-diego-persistence

gcloud compute addresses create ${CLUSTER_IP_NAME} --region us-west1 || true
LB_IP=$(gcloud compute addresses describe ${CLUSTER_IP_NAME} --region us-west1 --format json | jq -r '.address')

## DNS
gcloud dns managed-zones create ${CLUSTER_NAME} \
            --dns-name ${CLUSTER_DNS}. \
            --description "Managed zone for ${CLUSTER_NAME}" || true

NS0=$(gcloud dns managed-zones describe ${CLUSTER_NAME} --format json | jq -r '.nameServers[0]')
NS1=$(gcloud dns managed-zones describe ${CLUSTER_NAME} --format json | jq -r '.nameServers[1]')
NS2=$(gcloud dns managed-zones describe ${CLUSTER_NAME} --format json | jq -r '.nameServers[2]')
NS3=$(gcloud dns managed-zones describe ${CLUSTER_NAME} --format json | jq -r '.nameServers[3]')

gcloud dns record-sets transaction start --zone="${CLUSTER_NAME}"
gcloud dns record-sets transaction add ${LB_IP} --name="*.${CLUSTER_NAME}.persi.cf-app.com" \
  --ttl="30" \
  --type="A" \
  --zone="${CLUSTER_NAME}"
gcloud dns record-sets transaction execute --zone="${CLUSTER_NAME}"
## DNS

## DNS AWS
# aws create route53 with the ${ns} from above
cat << EOF > file.txt
{
  "Comment": "deploy k8s cluster creating route53 record",
  "Changes": [
    {
      "Action": "UPSERT",
      "ResourceRecordSet": {
        "Name": "${CLUSTER_NAME}.persi.cf-app.com.",
        "Type": "NS",
        "TTL": 0,
        "ResourceRecords": [
          {"Value": "$NS0"},
          {"Value": "$NS1"},
          {"Value": "$NS2"},
          {"Value": "$NS3"}
        ]
      }
    }
  ]
}
EOF

## DNS AWS
aws route53 change-resource-record-sets \
  --hosted-zone-id /hostedzone/$AWS_HOSTED_ZONE \
  --change-batch "$(cat ./file.txt)"

# output mapping
mkdir -p cluster-info
cluster_info_json="{\"name\":\"${CLUSTER_NAME}\", \"zone\":\"${ZONE}\", \"cluster_ip_name\": \"${CLUSTER_IP_NAME}\", \"lb_ip\": \"${LB_IP}\"}"
echo ${cluster_info_json} > cluster-info/cluster.json

pushd cf-for-k8s
    ./hack/generate-values.sh -d ${CLUSTER_DNS} > /tmp/cf-values.yml

    parsed_gcp_key=$(echo ${GCP_KEY} | tr -d '\r')
    cat << EOF >> /tmp/cf-values.yml

app_registry:
  hostname: gcr.io
  repository: gcr.io/cff-diego-persistence/cf-workloads
  username: _json_key
  password: '${GCP_KEY}'

istio_static_ip: ${LB_IP}
EOF
    ytt -f config -f /tmp/cf-values.yml > /tmp/cf-for-k8s-rendered.yml
    kapp deploy -a cf -f /tmp/cf-for-k8s-rendered.yml -y
popd

pushd smb-volume-k8s-release
    kubectl get namespace cf-smb || kubectl create namespace cf-smb
    kubectl get namespace cf-workloads || kubectl create namespace cf-workloads

    pushd smb-broker
        kapp deploy -y -a smb-broker -f <(ytt -f ytt/ -v smbBrokerUsername=foo -v smbBrokerPassword=foo -v image.tag=latest)
    popd
    pushd smb-csi-driver
        kapp deploy -y -a smb-csi-driver -f <(ytt -f ytt/base -v image.tag=latest)
    popd
popd

# add cf-values to output mapping
cp /tmp/cf-values.yml cluster-info
