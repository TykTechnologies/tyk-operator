package integration

import (
	"time"
)

func retryOperation(timeout time.Duration, interval time.Duration, fn func() error) error {
	intervalTick := time.NewTicker(interval)
	timeoutTick := time.NewTicker(timeout)

	var err error

	if fn() == nil {
		return nil
	}

	select {
	case <-intervalTick.C:
		if err = fn(); err == nil {
			return nil
		}

	case <-timeoutTick.C:
		return err
	}

	return nil
}
