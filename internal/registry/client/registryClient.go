package registry

import (
	"context"
	"time"

	reg "github.com/uromahn/k8s-svc-registry/api/registry"
	"google.golang.org/grpc"
)

type ServiceRegistryClient struct {
	conn    *grpc.ClientConn
	timeout time.Duration
}

func NewServiceRegistryClient(conn *grpc.ClientConn, timeout time.Duration) ServiceRegistryClient {
	return ServiceRegistryClient{
		conn:    conn,
		timeout: timeout,
	}
}

func (c ServiceRegistryClient) Register(ctx context.Context, svcInfo *reg.ServiceInfo) (*reg.RegistrationResult, error) {
	client := reg.NewServiceRegistryClient(c.conn)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	result, err := client.Register(ctx, svcInfo)
	return result, err
}

func (c ServiceRegistryClient) UnRegister(ctx context.Context, svcInfo *reg.ServiceInfo) (*reg.RegistrationResult, error) {
	client := reg.NewServiceRegistryClient(c.conn)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	result, err := client.UnRegister(ctx, svcInfo)
	return result, err
}
