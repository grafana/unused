package clicommon

import (
	"flag"
)

// ProviderFlags adds the provider configuration flags to the given
// flag set.
func ProviderFlags(fs *flag.FlagSet, gcp, aws, azure *StringSliceFlag) {
	fs.Var(gcp, "gcp.project", "GCP project ID (can be specified multiple times)")
	fs.Var(aws, "aws.profile", "AWS profile (can be specified multiple times)")
	fs.Var(azure, "azure.sub", "Azure subscription (can be specified multiple times)")
}
