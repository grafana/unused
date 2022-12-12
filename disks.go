package unused

import "sort"

// Disks is a collection of Disk.
type Disks []Disk

// ByFunc is the type of sorting functions for Disks.
type ByFunc func(p, q Disk) bool

// ByProvider sorts a Disks collection by provider name.
func ByProvider(p, q Disk) bool {
	return p.Provider().Name() < q.Provider().Name()
}

// ByName sorts a Disks collection by disk name.
func ByName(p, q Disk) bool {
	return p.Name() < q.Name()
}

// ByCreatedAt sorts a Disks collection by disk creation time.
func ByCreatedAt(p, q Disk) bool {
	return p.CreatedAt().Before(q.CreatedAt())
}

// Sort sorts the collection by the given function.
func (d Disks) Sort(by ByFunc) {
	sort.Slice(d, func(i, j int) bool { return by(d[i], d[j]) })
}
