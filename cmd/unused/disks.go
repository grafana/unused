package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/grafana/unused"
)

func listUnusedDisks(ctx context.Context, providers []unused.Provider) (unused.Disks, error) {
	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		total unused.Disks
	)

	wg.Add(len(providers))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, len(providers))

	for _, p := range providers {
		go func(p unused.Provider) {
			defer wg.Done()

			disks, err := p.ListUnusedDisks(ctx)
			if err != nil {
				cancel()
				errCh <- fmt.Errorf("%s %s: %w", p.Name(), p.Meta(), err)
				return
			}

			mu.Lock()
			total = append(total, disks...)
			mu.Unlock()
		}(p)
	}

	wg.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
	}

	return total, nil
}
