package upgrade

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"nts/vanguard/internal/hardware"
	"nts/vanguard/internal/store"
)

const (
	waitPIDTimeout  = 60 * time.Second
	waitLockTimeout = 30 * time.Second
	defaultHealth   = 8 * time.Second
)

type ApplyOpts struct {
	PID     int
	Target  string
	From    string
	Version string
	LockDir string
	Health  time.Duration
}

func (o ApplyOpts) paths() (Paths, error) {
	p, err := ForExecutable(o.Target)
	if err != nil {
		return Paths{}, err
	}
	if o.From != "" {
		p.New = o.From
	}
	return p, nil
}

// Install downloads the version from FTP, verifies sha256, stages exe.new,
// and starts a detached updater. The caller must then exit this process.
func Install(st store.Store, ver, lockDir string) error {
	p, err := Current()
	if err != nil {
		return err
	}
	_ = os.RemoveAll(p.Staging)
	if err := Download(st, ver, p.Staging); err != nil {
		return err
	}
	if err := Stage(p, ver); err != nil {
		return err
	}
	return StartUpdater(p, ver, os.Getpid(), lockDir)
}

// Stage copies the verified payload to exe.new and writes status=ready.
func Stage(p Paths, ver string) error {
	payload, hashFile, err := FindPayload(p.Staging, filepath.Base(p.Exe))
	if err != nil {
		return err
	}
	sumb, err := os.ReadFile(hashFile)
	if err != nil {
		return err
	}
	sum, err := ParseHashFile(sumb)
	if err != nil {
		return err
	}
	if err := VerifyFile(payload, sum); err != nil {
		return err
	}
	if err := copyFile(payload, p.New); err != nil {
		return err
	}
	if err := VerifyFile(p.New, sum); err != nil {
		_ = os.Remove(p.New)
		return err
	}
	if err := WriteStatus(p.Status, Status{Version: ver, State: StateReady}); err != nil {
		return err
	}
	_ = os.RemoveAll(p.Staging)
	return nil
}

func StartUpdater(p Paths, ver string, pid int, lockDir string) error {
	SetLogDir(p.Dir)
	self, err := os.Executable()
	if err != nil {
		return err
	}
	tmp := filepath.Join(os.TempDir(), fmt.Sprintf("vanguard-updater-%d%s", pid, exeSuffix()))
	if err := copyFile(self, tmp); err != nil {
		return err
	}
	args := []string{
		"--apply-update",
		"--pid", strconv.Itoa(pid),
		"--target", p.Exe,
		"--from", p.New,
		"--version", ver,
		"--lock-dir", lockDir,
	}
	logf("starting updater %s %v", tmp, args)
	return startDetached(tmp, args)
}

// Apply is the updater process: wait for the old POS, swap files, launch,
// and roll back if the new process dies immediately.
func Apply(opts ApplyOpts) error {
	if opts.Health <= 0 {
		opts.Health = defaultHealth
	}
	p, err := opts.paths()
	if err != nil {
		return err
	}
	logf("apply version=%s pid=%d target=%s from=%s", opts.Version, opts.PID, p.Exe, p.New)
	if err := waitPID(opts.PID, waitPIDTimeout); err != nil {
		return fmt.Errorf("wait pid: %w", err)
	}
	if err := hardware.WaitLockReleased(opts.LockDir, waitLockTimeout); err != nil {
		return fmt.Errorf("wait lock: %w", err)
	}
	if _, err := os.Stat(p.New); err != nil {
		return fmt.Errorf("staged file: %w", err)
	}
	if err := swap(p); err != nil {
		_ = WriteStatus(p.Status, Status{Version: opts.Version, State: StateFailed})
		return err
	}
	if err := WriteStatus(p.Status, Status{Version: opts.Version, State: StateApplied}); err != nil {
		_ = restore(p)
		return err
	}
	proc, err := startApp(p.Exe)
	if err != nil {
		logf("start new exe failed: %v", err)
		if rerr := restore(p); rerr != nil {
			logf("restore after start failure: %v", rerr)
		}
		_ = WriteStatus(p.Status, Status{Version: opts.Version, State: StateFailed})
		_, _ = startApp(p.Exe)
		return err
	}
	time.Sleep(opts.Health)
	if pidAlive(proc.Pid) {
		logf("new process %d still running", proc.Pid)
		return nil
	}
	logf("new process %d exited, rolling back", proc.Pid)
	if err := restore(p); err != nil {
		_ = WriteStatus(p.Status, Status{Version: opts.Version, State: StateFailed})
		return err
	}
	_ = WriteStatus(p.Status, Status{Version: opts.Version, State: StateFailed})
	_, err = startApp(p.Exe)
	return err
}

func swap(p Paths) error {
	_ = os.Remove(p.Bak)
	if err := os.Rename(p.Exe, p.Bak); err != nil {
		return fmt.Errorf("rename current to bak: %w", err)
	}
	if err := os.Rename(p.New, p.Exe); err != nil {
		_ = os.Rename(p.Bak, p.Exe)
		return fmt.Errorf("rename new to current: %w", err)
	}
	return nil
}

func restore(p Paths) error {
	bad := p.Exe + ".bad"
	_ = os.Remove(bad)
	if _, err := os.Stat(p.Exe); err == nil {
		_ = os.Rename(p.Exe, bad)
	}
	if err := os.Rename(p.Bak, p.Exe); err != nil {
		return fmt.Errorf("restore bak: %w", err)
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	return err
}

func logf(format string, args ...any) {
	f, err := os.OpenFile(LogPath(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		f, err = os.OpenFile(filepath.Join(os.TempDir(), appLogName), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return
		}
	}
	defer f.Close()
	_, _ = fmt.Fprintf(f, time.Now().Format("2006-01-02 15:04:05")+" "+format+"\n", args...)
	_ = f.Sync()
}
