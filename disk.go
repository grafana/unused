package unused

import "time"

// Disk represents an unused disk on a given cloud provider.
type Disk interface {
	// ID should return a unique ID for a disk within each cloud
	// provider.
	ID() string

	// Provider returns a reference to the provider used to instantiate
	// this disk
	Provider() Provider

	// Name returns the disk name.
	Name() string

	// SizeGB returns the disk size in GB (Azure/GCP) and GiB for AWS.
	SizeGB() int

	// SizeBytes returns the disk size in bytes.
	SizeBytes() float64

	// CreatedAt returns the time when the disk was created.
	CreatedAt() time.Time

	// LastUsedAt returns the date when the disk was last used.
	LastUsedAt() time.Time

	// Meta returns the disk metadata.
	Meta() Meta

	// DiskType returns the normalized type of disk.
	DiskType() DiskType
}

type DiskType string

const (
	SSD     DiskType = "ssd"
	HDD     DiskType = "hdd"
	Unknown DiskType = "unknown"
)

const GiBbytes = 1_073_741_824 // 2^30
