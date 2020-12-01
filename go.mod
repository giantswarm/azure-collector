module github.com/giantswarm/azure-collector/v2

go 1.15

require (
	github.com/Azure/azure-sdk-for-go v45.1.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.12
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.3
	github.com/Azure/go-autorest/autorest/to v0.4.0
	github.com/Azure/go-autorest/autorest/validation v0.3.0 // indirect
	github.com/giantswarm/apiextensions/v2 v2.6.2
	github.com/giantswarm/apiextensions/v3 v3.9.0
	github.com/giantswarm/exporterkit v0.2.0
	github.com/giantswarm/k8sclient/v4 v4.0.0
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.2.1
	github.com/giantswarm/microkit v0.2.2
	github.com/giantswarm/micrologger v0.4.0
	github.com/giantswarm/operatorkit/v2 v2.0.2
	github.com/giantswarm/statusresource/v2 v2.0.0
	github.com/giantswarm/versionbundle v0.2.0
	github.com/google/go-cmp v0.5.4
	github.com/prometheus/client_golang v1.8.0
	github.com/spf13/viper v1.7.1
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	k8s.io/api v0.18.9
	k8s.io/apimachinery v0.18.9
	k8s.io/client-go v0.18.9
	sigs.k8s.io/cluster-api v0.3.11
	sigs.k8s.io/controller-runtime v0.6.4
)

replace (
	github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.24+incompatible
	sigs.k8s.io/cluster-api => github.com/giantswarm/cluster-api v0.3.10-gs
)
