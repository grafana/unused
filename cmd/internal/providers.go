package internal

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"

	azcompute "github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/grafana/unused"
	"github.com/grafana/unused/aws"
	"github.com/grafana/unused/azure"
	"github.com/grafana/unused/gcp"
	"google.golang.org/api/compute/v1"
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

		p, err := aws.NewProvider(ec2.NewFromConfig(cfg), map[string]string{"profile": profile})
		if err != nil {
			return nil, fmt.Errorf("creating AWS provider for profile %s: %w", profile, err)
		}
		providers = append(providers, p)
	}

	if len(azureSubs) > 0 {
		a, err := auth.NewAuthorizerFromCLI()
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
func ProviderFlags(fs *flag.FlagSet, gcp, aws, azure *StringSliceFlag) {
	fs.Var(gcp, "gcp.project", "GCP project ID (can be specified multiple times)")
	fs.Var(aws, "aws.profile", "AWS profile (can be specified multiple times)")
	fs.Var(azure, "azure.sub", "Azure subscription (can be specified multiple times)")
}
