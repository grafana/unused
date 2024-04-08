package azure

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/grafana/unused"
)

var _ unused.Disk = &Disk{}

// Disk holds information about an Azure compute disk.
type Disk struct {
	compute.Disk
	provider *Provider
	meta     unused.Meta
}

// ID returns the Azure compute disk ID.
func (d *Disk) ID() string { return *d.Disk.ID }

// Provider returns a reference to the provider used to instantiate
// this disk.
func (d *Disk) Provider() unused.Provider { return d.provider }

// Name returns the name of this Azure compute disk.
func (d *Disk) Name() string { return *d.Disk.Name }

// SizeGB returns the size of this Azure compute disk in GB.
func (d *Disk) SizeGB() int { return int(*d.Disk.DiskSizeGB) }

// SizeBytes returns the size of this Azure compute disk in bytes.
func (d *Disk) SizeBytes() int { return int(*d.Disk.DiskSizeBytes) }

// CreatedAt returns the time when this Azure compute disk was
// created.
func (d *Disk) CreatedAt() time.Time { return d.Disk.TimeCreated.ToTime() }

// Meta returns the disk metadata.
func (d *Disk) Meta() unused.Meta { return d.meta }

// LastUsedAt returns a zero [time.Time] value, as Azure does not
// provide this information.
func (d *Disk) LastUsedAt() time.Time { return time.Time{} }

// DiskType Type returns the type of this Azure compute disk.
func (d *Disk) DiskType() unused.DiskType {
	switch d.Disk.Sku.Name {
	case compute.StandardLRS:
		return unused.HDD
	case compute.PremiumLRS, compute.StandardSSDLRS, compute.UltraSSDLRS:
		return unused.SSD
	default:
		return unused.Unknown
	}
}
