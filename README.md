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


