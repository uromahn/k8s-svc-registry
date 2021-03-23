package servertypes

import (
	"context"

	reg "github.com/uromahn/k8s-svc-registry/api/registry"
)

type ResultMsg struct {
	Result *reg.RegistrationResult
	Err    error
}

type RegOperation int

const (
	Register RegOperation = iota
	Unregister
)

type RegistrationMsg struct {
	Ctx             context.Context
	ResponseChannel chan ResultMsg
	SvcInfo         *reg.ServiceInfo
	Op              RegOperation
	Retry           bool
}
