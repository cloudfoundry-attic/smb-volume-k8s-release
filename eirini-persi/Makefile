build:
	kubectl kustomize ./base | kubectl apply -f -

delete:
	kubectl kustomize ./base | kubectl delete -f -

.PHONY: build