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

	// LastUsedAt returns the date when the disk was last mounted, or
	// time.Zero if this information isn't available (i.e. limited in
	// AWS and non-existing in Azure)
	LastUsedAt() time.Time

	// Meta returns the disk metadata
	Meta() Meta
}
