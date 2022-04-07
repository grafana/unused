package unused

import "context"

// Provider represents a cloud provider
type Provider interface {
	// Name returns the provider name
	Name() string

	// ListUnusedDisks returns a list of unused disks for the viden provider
	ListUnusedDisks(ctx context.Context) (Disks, error)
}
