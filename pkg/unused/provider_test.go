package unused

import "context"

type provider string

func (p *provider) Name() string { return string(*p) }

func (p *provider) ListUnusedDisks(ctx context.Context) (Disks, error) {
	return nil, nil
}
