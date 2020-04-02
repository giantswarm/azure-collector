package installation

import (
	"github.com/giantswarm/azure-collector/flag/service/installation/tenant"
)

type Installation struct {
	Name   string
	Tenant tenant.Tenant
}
