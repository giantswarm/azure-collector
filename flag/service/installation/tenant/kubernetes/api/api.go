package api

import (
	"github.com/giantswarm/azure-collector/flag/service/installation/tenant/kubernetes/api/auth"
)

type API struct {
	Auth auth.Auth
}
