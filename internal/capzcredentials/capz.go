package capzcredentials

import (
	"context"

	"github.com/giantswarm/microerror"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetAzureCredentialsFromMetadata(ctx context.Context, ctrlClient client.Client, obj metav1.ObjectMeta) (*AzureCredentials, error) {
	azureCredentials, err := getCapzCredentials(ctx, ctrlClient, obj)
	if IsMissingIdentityRef(err) || errors.IsNotFound(err) {
		// Unable to find the Identity Ref or one of the related resources.
		// We need to fall back to the organization logic to retrieve credentials for azure API.
		azureCredentials, err = getLegacyCredentials(ctx, ctrlClient, obj)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	return azureCredentials, nil
}

func getCapzCredentials(ctx context.Context, ctrlClient client.Client, obj metav1.ObjectMeta) (*AzureCredentials, error) {
	azureCluster, err := getAzureClusterFromMetadata(ctx, ctrlClient, obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if azureCluster.Spec.IdentityRef == nil {
		return nil, microerror.Maskf(missingIdentityRefError, "IdentiyRef was nil in AzureCluster %s/%s", azureCluster.Namespace, azureCluster.Name)
	}

	identity := capz.AzureClusterIdentity{}
	err = ctrlClient.Get(ctx, client.ObjectKey{Namespace: azureCluster.Spec.IdentityRef.Namespace, Name: azureCluster.Spec.IdentityRef.Name}, &identity)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	secret := v1.Secret{}
	err = ctrlClient.Get(ctx, client.ObjectKey{Namespace: identity.Spec.ClientSecret.Namespace, Name: identity.Spec.ClientSecret.Name}, &secret)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return &AzureCredentials{
		SubscriptionID: azureCluster.Spec.SubscriptionID,
		TenantID:       identity.Spec.TenantID,
		ClientID:       identity.Spec.ClientID,
		ClientSecret:   string(secret.Data["clientSecret"]),
	}, nil
}

func getAzureClusterFromMetadata(ctx context.Context, c client.Client, obj metav1.ObjectMeta) (*capz.AzureCluster, error) {
	// Check if "cluster.x-k8s.io/cluster-name" label is set.
	if obj.Labels[capi.ClusterLabelName] == "" {
		err := microerror.Maskf(invalidObjectMetaError, "Label %q must not be empty for object %q", capi.ClusterLabelName, obj.GetSelfLink())
		return nil, microerror.Mask(err)
	}

	return getAzureClusterByName(ctx, c, obj.Namespace, obj.Labels[capi.ClusterLabelName])
}

func getAzureClusterByName(ctx context.Context, c client.Client, namespace, name string) (*capz.AzureCluster, error) {
	azureCluster := &capz.AzureCluster{}
	key := client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}

	if err := c.Get(ctx, key, azureCluster); err != nil {
		return nil, microerror.Mask(err)
	}

	return azureCluster, nil
}
