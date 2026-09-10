//go:build !windows

package hardware

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// WaitLockReleased waits until the single-instance flock is free.
func WaitLockReleased(dir string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, "app.lock")
	deadline := time.Now().Add(timeout)
	for {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
		if err != nil {
			return err
		}
		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
			_ = f.Close()
			return nil
		}
		_ = f.Close()
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for application lock")
		}
		time.Sleep(200 * time.Millisecond)
	}
}
