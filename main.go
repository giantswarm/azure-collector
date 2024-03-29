package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/microkit/command"
	microserver "github.com/giantswarm/microkit/server"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/versionbundle"
	"github.com/spf13/viper"

	"github.com/giantswarm/azure-collector/v3/pkg/project"

	"github.com/giantswarm/azure-collector/v3/flag"
	"github.com/giantswarm/azure-collector/v3/server"
	"github.com/giantswarm/azure-collector/v3/service"
)

var (
	f = flag.New()
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	err := mainError()
	if err != nil {
		panic(fmt.Sprintf("%#v\n", err))
	}
}

func mainError() error {
	var err error

	ctx := context.Background()
	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		return microerror.Mask(err)
	}

	// We define a server factory to create the custom server once all command
	// line flags are parsed and all microservice configuration is sorted out.
	serverFactory := func(v *viper.Viper) microserver.Server {
		// Create a new custom service which implements business logic.
		var newService *service.Service
		{
			c := service.Config{
				Flag:   f,
				Logger: logger,
				Viper:  v,

				Description: project.Description(),
				GitCommit:   project.GitSHA(),
				ProjectName: project.Name(),
				Source:      project.Source(),
				Version:     project.Version(),
			}

			newService, err = service.New(c)
			if err != nil {
				panic(fmt.Sprintf("%#v", microerror.Mask(err)))
			}

			go newService.Boot(ctx)
		}

		// Create a new custom server which bundles our endpoints.
		var newServer microserver.Server
		{
			c := server.Config{
				Logger:  logger,
				Service: newService,
				Viper:   v,

				ProjectName: project.Name(),
			}

			newServer, err = server.New(c)
			if err != nil {
				panic(fmt.Sprintf("%#v", microerror.Mask(err)))
			}
		}

		return newServer
	}

	// Create a new microkit command which manages our custom microservice.
	var newCommand command.Command
	{
		c := command.Config{
			Logger:        logger,
			ServerFactory: serverFactory,

			Description:    project.Description(),
			GitCommit:      project.GitSHA(),
			Name:           project.Name(),
			Source:         project.Source(),
			Version:        project.Version(),
			VersionBundles: []versionbundle.Bundle{project.NewVersionBundle()},
		}

		newCommand, err = command.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	daemonCommand := newCommand.DaemonCommand().CobraCommand()

	daemonCommand.PersistentFlags().String(f.Service.Azure.ClientID, "", "ID of the Active Directory Service Principal.")
	daemonCommand.PersistentFlags().String(f.Service.Azure.ClientSecret, "", "Secret of the Active Directory Service Principal.")
	daemonCommand.PersistentFlags().String(f.Service.Azure.PartnerID, "", "Partner id used in Azure for the attribution partner program.")
	daemonCommand.PersistentFlags().String(f.Service.Azure.SubscriptionID, "", "ID of the Azure Subscription.")
	daemonCommand.PersistentFlags().String(f.Service.Azure.TenantID, "", "ID of the Active Directory Tenant.")
	daemonCommand.PersistentFlags().String(f.Service.ControlPlaneResourceGroup, "", "Control plane resource group name.")
	daemonCommand.PersistentFlags().String(f.Service.Location, "westeurope", "Azure location of the host and guset clusters.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.Address, "", "Address used to connect to Kubernetes. When empty in-cluster config is created.")
	daemonCommand.PersistentFlags().Bool(f.Service.Kubernetes.InCluster, true, "Whether to use the in-cluster config to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.KubeConfig, "", "KubeConfig used to connect to Kubernetes. When empty other settings are used.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.KubeConfigPath, "", "Optional path to KubeConfig file to connect to Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CAFile, "", "Certificate authority file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CrtFile, "", "Certificate file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.KeyFile, "", "Key file path to use to authenticate with Kubernetes.")

	err = newCommand.CobraCommand().Execute()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
