//go:build windows

package hardware

import (
	"fmt"
	"time"

	"golang.org/x/sys/windows"
)

const mutexName = "Local\\NTSVanguardPOS"

// WaitLockReleased waits until the single-instance mutex is free, then
// releases it so the replacement process can acquire it.
func WaitLockReleased(_ string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	name, err := windows.UTF16PtrFromString(mutexName)
	if err != nil {
		return err
	}
	h, err := windows.CreateMutex(nil, false, name)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	ms := uint32(timeout.Milliseconds())
	ev, err := windows.WaitForSingleObject(h, ms)
	if err != nil {
		return err
	}
	if ev != windows.WAIT_OBJECT_0 && ev != windows.WAIT_ABANDONED {
		return fmt.Errorf("timeout waiting for application lock")
	}
	_ = windows.ReleaseMutex(h)
	return nil
}
