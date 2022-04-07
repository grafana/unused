package unused

import (
	"fmt"
	"time"
)

var _ Disk = disk{}

type disk struct {
	name      string
	provider  *provider
	createdAt time.Time
}

func (d disk) Provider() Provider   { return d.provider }
func (d disk) Name() string         { return d.name }
func (d disk) CreatedAt() time.Time { return d.createdAt }

func (d disk) String() string {
	return fmt.Sprintf("disk{Provider:%q, Name:%q, CreatedAt:%q}",
		d.provider.Name(),
		d.name,
		d.createdAt.Format(time.RFC3339))
}
