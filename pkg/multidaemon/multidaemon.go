package multidaemon

import "context"

type DaemonFunc func(stopCh <-chan struct{}) error

func InitializeDaemons(signalStopCh <-chan struct{}, daemons ...DaemonFunc) []error {
	ctx, cancel := context.WithCancel(context.Background())
	stopCh := ctx.Done()
	doneCh := make(chan error)

	go func() {
		<-signalStopCh
		cancel()
	}()

	for _, o := range daemons {
		f := o
		go func() {
			err := f(stopCh)
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
