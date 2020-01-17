build:
	kubectl kustomize ./base | kubectl apply -f -

delete:
	kubectl delete mutatingwebhookconfiguration eirini-x-mutating-hook-eirini -n eirini || true
	kubectl delete secret eirini-x-setupcertificate -n eirini || true
	kubectl kustomize ./ephemeral-env | kubectl delete -f -

status:
	kubectl kustomize ./ephemeral-env | kubectl describe -f -

.PHONY: build delete status