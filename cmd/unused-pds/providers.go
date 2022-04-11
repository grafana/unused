package main

import (
	"context"
	"fmt"

	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/grafana/unused-pds/pkg/provider/aws"
	"github.com/grafana/unused-pds/pkg/provider/azure"
	"github.com/grafana/unused-pds/pkg/provider/gcp"
	"github.com/grafana/unused-pds/pkg/unused"
)

func createProviders(ctx context.Context, gcpProjects, awsProfiles, azureSubs []string) ([]unused.Provider, error) {
	providers := make([]unused.Provider, 0, len(gcpProjects)+len(awsProfiles)+len(azureSubs))

	for _, projectID := range gcpProjects {
		p, err := gcp.NewProvider(ctx, projectID, map[string]string{"project": projectID})
		if err != nil {
			return nil, fmt.Errorf("creating GCP provider for project %s: %w", projectID, err)
		}
		providers = append(providers, p)
	}

	for _, profile := range awsProfiles {
		p, err := aws.NewProvider(ctx, map[string]string{"profile": profile}, config.WithSharedConfigProfile(profile))
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
			p, err := azure.NewProvider(sub, azure.WithAuthorizer(a))
			if err != nil {
				return nil, fmt.Errorf("creating Azure provider for subscription %s: %w", sub, err)
			}
			providers = append(providers, p)
		}
	}

	return providers, nil
}
