package unused

import "context"

// Provider represents a cloud provider
type Provider interface {
	// ListUnusedDisks returns a list of unused disks for the viden provider
	ListUnusedDisks(ctx context.Context) (Disks, error)
}
