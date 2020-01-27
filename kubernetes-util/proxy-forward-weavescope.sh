
kubectl port-forward -n weave "$(kubectl get pod --selector=name=weave-scope-app -o jsonpath={.items..metadata.name} -n weave)" 4040
