kapp: SHELL:=/bin/bash
kapp:
	kubectl get namespace eirini || kubectl create namespace eirini
	kapp deploy -y -a eirini-persi -f <(ytt -f ytt/base -f ytt/overlay/ephemeral-env -v namespace=eirini)

delete:
	kubectl delete namespace eirini
	kubectl delete mutatingwebhookconfiguration eirini-x-mutating-hook-eirini -n eirini || true
	kubectl delete secret eirini-x-setupcertificate -n eirini || true
	kapp -y delete -a eirini-persi

status:
	kapp inspect -a eirini-persi

.PHONY: build delete status