package integration

import (
	"fmt"
	"time"
)

// retryOperation will retry the given 'fn' func at the given 'interval', until/unless 'timeout' is reached.
// Any non-nil errors returned by 'fn' in the interim will be ignored.
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
		return fmt.Errorf("Timeout: failed to complete operation:%v", err)
	}

	return nil
}
