package unused

import "time"

// Disk reprensents an unused disk on a given cloud provider
type Disk interface {
	// ID should return a unique ID for a disk within each cloud provider
	ID() string

	// Provider returns a string indicating the provider this disk belongs to
	Provider() Provider

	// Name returns the disk name
	Name() string

	// CreatedAt returns the time.Time when the disk was created
	CreatedAt() time.Time

	// Meta returns the disk metadata
	Meta() Meta
}
