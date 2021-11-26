# istio-falcon
Monitor istio Gateway (custom resource), auto expose istio-ingressgateway service port

## Overview
istio-ingressgateway service when use LoadBalance Or ExternalIps，default only expose 80,443 port. istio-falcon monitor Custom Resource Gateway,Automatically open external ports for istio-ingressgateway services.

istio-falcon use kubernetes leaseLeader ensure its Highly-Available.

## Install in Kubernetes
Use DaemonSet in master node
yaml file in manifest dir.
```shell
kubectl apply -f istio-falcon.yaml
```

## Run
istio-falcon mgr
- --in-cluster                 Use the inCluster token to authenticate to Kubernetes (default true)
- --kube-config string         k8s cluster kubeconfig file path
- --namespace string           istio-falcon pod run in namespace (default "kube-system")
- --ports strings              istio-ingressgateway service opened ports by default (default [80,443,15021])
- --service-name string        istio-ingressgateway service name (default "istio-ingressgateway")
- --service-namespace string   istio-ingressgateway service namespace (default "istio-system")


