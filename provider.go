package unused

import "context"

// Provider represents a cloud provider
type Provider interface {
	// Name returns the provider name
	Name() string

	// ListUnusedDisks returns a list of unused disks for the given provider
	ListUnusedDisks(ctx context.Context) (Disks, error)

	// Meta returns the provider metadata
	Meta() Meta

	// Delete deletes a disk from the provider
	Delete(ctx context.Context, disk Disk) error
}
