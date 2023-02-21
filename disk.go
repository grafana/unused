package unused

import "time"

// Disk represents an unused disk on a given cloud provider.
type Disk interface {
	// ID should return a unique ID for a disk within each cloud
	// provider.
	ID() string

	// Provider returns a string indicating the provider this disk
	// belongs to.
	Provider() Provider

	// Name returns the disk name.
	Name() string

	// SizeGB returns the disk size in GB.
	SizeGB() int

	// CreatedAt returns the time when the disk was created.
	CreatedAt() time.Time

	// LastUsedAt returns the date when the disk was last used.
	LastUsedAt() time.Time

	// Meta returns the disk metadata.
	Meta() Meta
	DiskType() string
}
