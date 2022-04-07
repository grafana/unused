package unused

import "sort"

// Disks is a collection of Disk
type Disks []Disk

type ByFunc func(p, q Disk) bool

func ByProvider(p, q Disk) bool {
	return p.Provider().Name() < q.Provider().Name()
}

func ByName(p, q Disk) bool {
	return p.Name() < q.Name()
}

func ByCreatedAt(p, q Disk) bool {
	return p.CreatedAt().Before(q.CreatedAt())
}

func (d Disks) Sort(by ByFunc) {
	sort.Slice(d, func(i, j int) bool { return by(d[i], d[j]) })
}
