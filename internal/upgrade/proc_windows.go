//go:build windows

package upgrade

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
)

func exeSuffix() string { return ".exe" }

func waitPID(pid int, timeout time.Duration) error {
	if pid <= 0 || !pidAlive(pid) {
		return nil
	}
	h, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		if !pidAlive(pid) {
			return nil
		}
		return err
	}
	defer windows.CloseHandle(h)
	ev, err := windows.WaitForSingleObject(h, uint32(timeout.Milliseconds()))
	if err != nil {
		return err
	}
	if ev == uint32(windows.WAIT_TIMEOUT) {
		return fmt.Errorf("timeout waiting for pid %d", pid)
	}
	return nil
}

func pidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer windows.CloseHandle(h)
	var code uint32
	if err := windows.GetExitCodeProcess(h, &code); err != nil {
		return false
	}
	return code == 259 // STILL_ACTIVE
}

func startDetached(name string, args []string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = filepath.Dir(name)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.DETACHED_PROCESS | windows.CREATE_NEW_PROCESS_GROUP | windows.CREATE_NO_WINDOW,
	}
	return cmd.Start()
}

func startApp(name string) (*os.Process, error) {
	cmd := exec.Command(name)
	cmd.Dir = filepath.Dir(name)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	go func() { _ = cmd.Wait() }()
	return cmd.Process, nil
}
