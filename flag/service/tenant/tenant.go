package tenant

import (
	"github.com/giantswarm/azure-collector/flag/service/tenant/ignition"
	"github.com/giantswarm/azure-collector/flag/service/tenant/ssh"
)

type Tenant struct {
	Ignition ignition.Ignition
	SSH      ssh.SSH
}
