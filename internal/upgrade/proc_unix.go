//go:build !windows

package upgrade

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

func exeSuffix() string { return "" }

func waitPID(pid int, timeout time.Duration) error {
	if pid <= 0 || !pidAlive(pid) {
		return nil
	}
	deadline := time.Now().Add(timeout)
	for pidAlive(pid) {
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for pid %d", pid)
		}
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}

func pidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	if err == nil || err == syscall.EPERM {
		return true
	}
	return false
}

func startDetached(name string, args []string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = filepath.Dir(name)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
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
