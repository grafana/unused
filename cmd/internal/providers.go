package internal

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	azcompute "github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/grafana/unused"
	"github.com/grafana/unused/aws"
	"github.com/grafana/unused/azure"
	"github.com/grafana/unused/gcp"
	compute "google.golang.org/api/compute/v1"
)

var ErrNoProviders = errors.New("please select at least one provider")

func CreateProviders(ctx context.Context, logger *slog.Logger, pc *ProviderConfig) ([]unused.Provider, error) {
	providers := make([]unused.Provider, 0, len(pc.GCPProjects)+len(pc.AWSProfiles)+len(pc.AzureSubs))

	for _, projectID := range pc.GCPProjects {
		svc, err := compute.NewService(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot create GCP compute service: %w", err)
		}
		p, err := gcp.NewProvider(logger, svc, pc.GCPProviderName, projectID, map[string]string{"project": projectID, "provider_name": "gke"})
		if err != nil {
			return nil, fmt.Errorf("creating GCP provider for project %s: %w", projectID, err)
		}
		providers = append(providers, p)
	}

	for _, profile := range pc.AWSProfiles {
		cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
		if err != nil {
			return nil, fmt.Errorf("cannot load AWS config for profile %s: %w", profile, err)
		}

		p, err := aws.NewProvider(logger, ec2.NewFromConfig(cfg), pc.AWSProviderName, map[string]string{"profile": profile})
		if err != nil {
			return nil, fmt.Errorf("creating AWS provider for profile %s: %w", profile, err)
		}
		providers = append(providers, p)
	}

	if len(pc.AzureSubs) > 0 {
		var a autorest.Authorizer
		var err error

		if os.Getenv("AZURE_CLIENT_ID") != "" && os.Getenv("AZURE_CLIENT_SECRET") != "" && os.Getenv("AZURE_TENANT_ID") != "" {
			a, err = auth.NewAuthorizerFromEnvironment()
		} else {
			a, err = auth.NewAuthorizerFromCLI()
		}
		if err != nil {
			return nil, fmt.Errorf("creating Azure authorizer: %w", err)
		}

		for _, sub := range pc.AzureSubs {
			c := azcompute.NewDisksClient(sub)
			c.Authorizer = a

			p, err := azure.NewProvider(c, pc.AWSProviderName, map[string]string{"subscription": sub, "provider_name": "aks"})
			if err != nil {
				return nil, fmt.Errorf("creating Azure provider for subscription %s: %w", sub, err)
			}
			providers = append(providers, p)
		}
	}

	if len(providers) == 0 {
		return nil, ErrNoProviders
	}

	return providers, nil
}

// ProviderFlags adds the provider configuration flags to the given
// flag set.
type ProviderConfig struct {
	GCPProjects       StringSliceFlag
	AWSProfiles       StringSliceFlag
	AzureSubs         StringSliceFlag
	GCPProviderName   string
	AWSProviderName   string
	AzureProviderName string
}

func ProviderFlags(fs *flag.FlagSet, pc *ProviderConfig) {
	fs.Var(&pc.GCPProjects, "gcp.project", "GCP project ID (can be specified multiple times)")
	fs.Var(&pc.AWSProfiles, "aws.profile", "AWS profile (can be specified multiple times)")
	fs.Var(&pc.AzureSubs, "azure.sub", "Azure subscription (can be specified multiple times)")
	fs.StringVar(&pc.GCPProviderName, "gcp.providername", gcp.DefaultProviderName, `GCP provider name to use, default: "GCP" (e.g. "GKE")`)
	fs.StringVar(&pc.AWSProviderName, "aws.providername", aws.DefaultProviderName, `AWS provider name to use, default: "AWS" (e.g. "EKS")`)
	fs.StringVar(&pc.AzureProviderName, "azure.providername", azure.DefaultProviderName, `Azure provider name to use, default: "Azure" (e.g. "AKS")`)
}
