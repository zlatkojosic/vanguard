package upgrade

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const appLogName = "vanguard.log"

var (
	logDirMu sync.Mutex
	logDir   string
)

func SetLogDir(dir string) {
	logDirMu.Lock()
	logDir = dir
	logDirMu.Unlock()
}

func LogDir() string {
	logDirMu.Lock()
	defer logDirMu.Unlock()
	if logDir != "" {
		return logDir
	}
	return defaultLogDir()
}

func LogPath() string {
	return filepath.Join(LogDir(), appLogName)
}

func defaultLogDir() string {
	if exe, err := os.Executable(); err == nil {
		if r, err := filepath.EvalSymlinks(exe); err == nil {
			exe = r
		}
		base := strings.ToLower(filepath.Base(exe))
		if !strings.HasPrefix(base, "vanguard-updater") {
			return filepath.Dir(exe)
		}
	}
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return os.TempDir()
}

func OpenAppLog() io.Closer {
	path := LogPath()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		path = filepath.Join(os.TempDir(), appLogName)
		f, err = os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return nil
		}
	}
	w := &fileLog{f: f}
	log.SetOutput(w)
	log.SetFlags(log.LstdFlags)
	log.Println("log file", path)
	return w
}

// fileLog writes to the log file first and Syncs. Stderr is best-effort:
// a Windows GUI exe often has a broken stderr, and MultiWriter would then
// skip the file entirely.
type fileLog struct {
	mu sync.Mutex
	f  *os.File
}

func (w *fileLog) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n, err := w.f.Write(p)
	_ = w.f.Sync()
	_, _ = os.Stderr.Write(p)
	return n, err
}

func (w *fileLog) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.f.Close()
}
