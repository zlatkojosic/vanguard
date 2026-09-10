//go:build windows

package hardware

import (
	"fmt"

	"golang.org/x/sys/windows"
)

func AcquireLock(dir string) (*Lock, error) {
	name, err := windows.UTF16PtrFromString("Local\\NTSVanguardPOS")
	if err != nil {
		return nil, err
	}
	h, err := windows.CreateMutex(nil, false, name)
	if err != nil {
		return nil, err
	}
	event, err := windows.WaitForSingleObject(h, 0)
	if err != nil || (event != windows.WAIT_OBJECT_0 && event != windows.WAIT_ABANDONED) {
		_ = windows.CloseHandle(h)
		return nil, fmt.Errorf("aplikacija je već pokrenuta")
	}
	return &Lock{release: func() {
		_ = windows.ReleaseMutex(h)
		_ = windows.CloseHandle(h)
	}}, nil
}
