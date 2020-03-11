build:
	kubectl create namespace eirini
	kubectl kustomize ./base | kapp -y deploy -a eirini-persi -f -

delete:
	kubectl delete mutatingwebhookconfiguration eirini-x-mutating-hook-eirini -n eirini || true
	kubectl delete secret eirini-x-setupcertificate -n eirini || true
	kapp -y delete -a eirini-persi

status:
	kapp inspect -a eirini-persi

.PHONY: build delete status