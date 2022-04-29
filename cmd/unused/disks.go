package main

import (
	"context"
	"fmt"
	"os"
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

	for _, p := range providers {
		go func(p unused.Provider) {
			defer wg.Done()

			disks, err := p.ListUnusedDisks(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s %s failed: %v\n", p.Name(), p.Meta(), err)
				return
			}

			mu.Lock()
			total = append(total, disks...)
			mu.Unlock()
		}(p)
	}

	wg.Wait()

	return total, nil
}
