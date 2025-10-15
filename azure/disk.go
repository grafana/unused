package azure

import (
	"time"

	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/grafana/unused"
)

var _ unused.Disk = &Disk{}

// Disk holds information about an Azure compute disk.
type Disk struct {
	*compute.Disk
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
// Note that Azure uses binary GB, aka, GiB
// https://learn.microsoft.com/en-us/dotnet/api/microsoft.azure.management.compute.models.datadisk.disksizegb?view=azure-dotnet-legacy
func (d *Disk) SizeGB() int { return int(*d.Disk.Properties.DiskSizeGB) }

// SizeBytes returns the size of this Azure compute disk in bytes.
func (d *Disk) SizeBytes() float64 { return float64(*d.Disk.Properties.DiskSizeBytes) }

// CreatedAt returns the time when this Azure compute disk was
// created.
func (d *Disk) CreatedAt() time.Time { return *d.Disk.Properties.TimeCreated }

// Meta returns the disk metadata.
func (d *Disk) Meta() unused.Meta { return d.meta }

// LastUsedAt returns the time when the Azure compute disk was last
// detached.
func (d *Disk) LastUsedAt() time.Time {
	if d.Disk.Properties.LastOwnershipUpdateTime == nil {
		// Special case: disk was created manually and never used,
		// return the creation time.
		return d.CreatedAt()
	}
	return *d.Disk.Properties.LastOwnershipUpdateTime
}

// DiskType Type returns the type of this Azure compute disk.
func (d *Disk) DiskType() unused.DiskType {
	switch *d.Disk.SKU.Name {
	case compute.DiskStorageAccountTypesStandardLRS:
		return unused.HDD
	case compute.DiskStorageAccountTypesPremiumLRS,
		compute.DiskStorageAccountTypesStandardSSDLRS,
		compute.DiskStorageAccountTypesUltraSSDLRS:
		return unused.SSD
	default:
		return unused.Unknown
	}
}
