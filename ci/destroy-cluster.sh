#!/bin/bash -eu

echo ${GCP_KEY} > /tmp/gcp.json

gcloud auth activate-service-account  --key-file /tmp/gcp.json
gcloud config set project cff-diego-persistence

CLUSTER_NAME=$(cat cluster-info/cluster.json| jq -r '.name')
ZONE=$(cat cluster-info/cluster.json| jq -r '.zone')
CLUSTER_IP_NAME=$(cat cluster-info/cluster.json| jq -r '.cluster_ip_name')
LB_IP=$(cat cluster-info/cluster.json| jq -r '.lb_ip')

gcloud -q compute addresses delete ${CLUSTER_IP_NAME} --region us-west1 || true

gcloud -q container clusters delete ${CLUSTER_NAME} --zone ${ZONE}

NS0=$(gcloud dns managed-zones describe ${CLUSTER_NAME} --format json | jq -r '.nameServers[0]')
NS1=$(gcloud dns managed-zones describe ${CLUSTER_NAME} --format json | jq -r '.nameServers[1]')
NS2=$(gcloud dns managed-zones describe ${CLUSTER_NAME} --format json | jq -r '.nameServers[2]')
NS3=$(gcloud dns managed-zones describe ${CLUSTER_NAME} --format json | jq -r '.nameServers[3]')

cat << EOF > file.txt
{
  "Comment": "deploy k8s cluster creating route53 record",
  "Changes": [
    {
      "Action": "DELETE",
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
aws route53 change-resource-record-sets \
  --hosted-zone-id /hostedzone/$AWS_HOSTED_ZONE \
  --change-batch "$(cat ./file.txt)"

gcloud dns record-sets transaction start --zone="${CLUSTER_NAME}"
gcloud dns record-sets transaction remove ${LB_IP} --name="*.${CLUSTER_NAME}.persi.cf-app.com" \
  --ttl="30" \
  --type="A" \
  --zone="${CLUSTER_NAME}"
gcloud dns record-sets transaction execute --zone="${CLUSTER_NAME}"
gcloud -q dns managed-zones delete ${CLUSTER_NAME}
