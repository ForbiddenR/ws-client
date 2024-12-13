package wait

import "golang.org/x/sync/errgroup"

type Starter interface {
	Start() error
}

func Start(starts ...Starter) error {
	eg := errgroup.Group{}
	for _, v := range starts {
		eg.Go(func() error {
			return v.Start()
		})
	}
	return eg.Wait()
}
