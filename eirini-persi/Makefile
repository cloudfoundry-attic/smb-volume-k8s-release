build:
	kubectl kustomize ./base | kubectl apply -f -

delete:
	kubectl delete mutatingwebhookconfiguration eirini-x-mutating-hook-eirini -n eirini || true
	kubectl delete secret eirini-x-setupcertificate -n eirini || true
	kubectl kustomize ./base | kubectl delete -f -

.PHONY: build delete