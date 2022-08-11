package retry

import "github.com/qingwave/gocorex/utils/wait"

func RetryOnError(backoff wait.Backoff, fn func() error) error {
	return RetryOnCondition(backoff, func(err error) bool {
		return err != nil
	}, fn)
}

func RetryOnCondition(backoff wait.Backoff, retriable func(error) bool, fn func() error) error {
	var lastErr error
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := fn()
		switch {
		case err == nil:
			return true, nil
		case retriable(err):
			lastErr = err
			return false, nil
		default:
			return false, err
		}
	})
	if err == wait.ErrWaitTimeout {
		err = lastErr
	}
	return err
}
