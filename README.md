# Kubernetes Service Registry

This is a gRPC server that receives registration requests for services outside of Kubernetes and registers their endpoints with a headless service inside K8s so they can be discovered by services inside the K8s cluster.

However this assumes that all pods inside K8s can access the endpoints outside of the K8s cluster. If that is not possible, an egress gateway will have to be installed and the service endpoints should be registered via a DNS name that can be resolved via the egress gateway.

Another update to test something in Git.