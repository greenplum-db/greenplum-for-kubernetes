package multidaemon

import "context"

type DaemonFunc func(ctx context.Context) error

func InitializeDaemons(signalCtx context.Context, daemons ...DaemonFunc) []error {
	ctx, cancel := context.WithCancel(context.Background())
	doneCh := make(chan error)

	go func() {
		<-signalCtx.Done()
		cancel()
	}()

	for _, o := range daemons {
		f := o
		go func() {
			err := f(ctx)
			if err != nil {
				cancel()
			}
			doneCh <- err
		}()
	}

	var errs []error
	for range daemons {
		if err := <-doneCh; err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
