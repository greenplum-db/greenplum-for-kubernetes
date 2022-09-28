package multihost

type Operation interface {
	Execute(addr string) error
}

// ParallelForeach performs an operation for every host in a hostname list (in parallel).
// On success, it returns an empty slice of errors.
// On failure, it returns a slice containing the accumulated errors (in no particular order).
func ParallelForeach(operation Operation, hostnames []string) []error {
	errorCh := make(chan error, len(hostnames))
	for _, hostname := range hostnames {
		go func(hostname string) {
			errorCh <- operation.Execute(hostname)
		}(hostname)
	}
	var errors []error
	for range hostnames {
		if err := <-errorCh; err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}
