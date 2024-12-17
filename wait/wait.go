package wait

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type Starter interface {
	Start(context.Context) error
}

func Start(starts ...Starter) error {
	eg, errCtx := errgroup.WithContext(context.Background())
	for _, v := range starts {
		eg.Go(func() error {
			return v.Start(errCtx)
		})
	}
	return eg.Wait()
}
