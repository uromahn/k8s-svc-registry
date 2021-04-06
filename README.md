# Kubernetes Service Registry

This is a gRPC server that receives registration requests for services outside of Kubernetes and registers their endpoints with a headless service inside K8s so they can be discovered by services inside the K8s cluster.

However this assumes that all pods inside K8s can access the endpoints outside of the K8s cluster. If that is not possible, an egress gateway will have to be installed and the service endpoints should be registered via a DNS name that can be resolved via the egress gateway.

## Implementation Notes

We are currently using two methods:
```
rpc Register (ServiceInfo) returns (RegistrationResult) {}
```
and
```
rpc UnRegister (ServiceInfo) returns (RegistrationResult) {}
```

However, we want to add an additional function that allows bi-directional streaming:
```
rpc RegistrationEvent (stream RegistrationMsg) returns (stream RegistrationResponse) {}
```

As a result of that, the current `ServiceInfo` message should be changed as follows:
```
enum RegistrationType {
    REGISTER = 0;
    UNREGISTER = 1;
}
message RegistrationMsg {
    // Type of registration event
    RegistrationType type = 1;
    // namespace the service is registered in
    string namespace = 2;
    // name of the service - this determines which namespace this service uses
    string service_name = 3;
    // DNS name of the service host not including subdomain or domain and no dots
    string host_name = 4;
    // IP address of that specific host name
    string ipaddress = 5;
    // Physical node name where the service instance is running.
    // For bare metal or VM services, this may the the FQN of the service including the host_name
    string node_name = 6;
    // Array of named ports for the service
    repeated NamedPort ports = 7;
    // A weight this service instance should be given
    // NOTE: this information may not be utilized by the consumer of this information
    float weight = 8;
}
```
and a new `RegistrationResponse` should look like this
```
message RegistrationResponse {
    // Type of registration event
    RegistrationType type = 1;
    // namespace the service is registered in
    string namespace = 2;
    // name of the service
    string service_name = 3;
    // IP address of that specific host name
    string ipaddress = 4;
    // Array of named ports for the service
    repeated NamedPort ports = 5;
    // Status code of the operation
    // We will reuse the HTTP status code, e.g. a 201 will signal that the entry
    // has been successfully created.
    uint32 status = 6;
    // Details message explaining the status
    string status_details = 7;
}
```
