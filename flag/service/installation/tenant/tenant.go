package tenant

import (
	"github.com/giantswarm/azure-collector/flag/service/installation/tenant/kubernetes"
)

type Tenant struct {
	Kubernetes kubernetes.Kubernetes
}
