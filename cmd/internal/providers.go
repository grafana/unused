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

func CreateProviders(ctx context.Context, logger *slog.Logger, gcpProjects, awsProfiles, azureSubs []string) ([]unused.Provider, error) {
	providers := make([]unused.Provider, 0, len(gcpProjects)+len(awsProfiles)+len(azureSubs))

	for _, projectID := range gcpProjects {
		svc, err := compute.NewService(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot create GCP compute service: %w", err)
		}
		p, err := gcp.NewProvider(logger, svc, projectID, map[string]string{"project": projectID})
		if err != nil {
			return nil, fmt.Errorf("creating GCP provider for project %s: %w", projectID, err)
		}
		providers = append(providers, p)
	}

	for _, profile := range awsProfiles {
		cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
		if err != nil {
			return nil, fmt.Errorf("cannot load AWS config for profile %s: %w", profile, err)
		}

		p, err := aws.NewProvider(logger, ec2.NewFromConfig(cfg), map[string]string{"profile": profile})
		if err != nil {
			return nil, fmt.Errorf("creating AWS provider for profile %s: %w", profile, err)
		}
		providers = append(providers, p)
	}

	if len(azureSubs) > 0 {
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

		for _, sub := range azureSubs {
			c := azcompute.NewDisksClient(sub)
			c.Authorizer = a

			p, err := azure.NewProvider(c, map[string]string{"subscription": sub})
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
func ProviderFlags(fs *flag.FlagSet, gcpProject, awsProfile, azureSub *StringSliceFlag) {
	fs.Var(gcpProject, "gcp.project", "GCP project ID (can be specified multiple times)")
	fs.Var(awsProfile, "aws.profile", "AWS profile (can be specified multiple times)")
	fs.Var(azureSub, "azure.sub", "Azure subscription (can be specified multiple times)")
	fs.StringVar(&gcp.ProviderName, "gcp.providername", gcp.ProviderName, `GCP provider name to use, default: "GCP" (e.g. "GKE")`)
	fs.StringVar(&aws.ProviderName, "aws.providername", aws.ProviderName, `AWS provider name to use, default: "AWS" (e.g. "EKS")`)
	fs.StringVar(&azure.ProviderName, "azure.providername", azure.ProviderName, `Azure provider name to use, default: "Azure" (e.g. "AKS")`)
}
