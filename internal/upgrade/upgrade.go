package upgrade

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"nts/vanguard/internal/store"
)

const (
	ftpUser = "application"
	ftpPass = "123app!"
	ftpRoot = "vanguard/versions"

	statusName = "vanguard.upgrade"
	stagingDir = "newVersion"
)

const (
	StateReady   = "ready"
	StateApplied = "applied"
	StateFailed  = "failed"
)

// Paths are files next to the running POS binary.
type Paths struct {
	Exe     string
	Dir     string
	New     string
	Bak     string
	Staging string
	Status  string
}

type Status struct {
	Version string `json:"version"`
	State   string `json:"state"`
}

func ForExecutable(exe string) (Paths, error) {
	if exe == "" {
		return Paths{}, fmt.Errorf("empty executable path")
	}
	exe, err := filepath.Abs(exe)
	if err != nil {
		return Paths{}, err
	}
	dir := filepath.Dir(exe)
	return Paths{
		Exe:     exe,
		Dir:     dir,
		New:     exe + ".new",
		Bak:     exe + ".bak",
		Staging: filepath.Join(dir, stagingDir),
		Status:  filepath.Join(dir, statusName),
	}, nil
}

func Current() (Paths, error) {
	exe, err := os.Executable()
	if err != nil {
		return Paths{}, err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return ForExecutable(exe)
}

func ReadStatus(path string) Status {
	b, err := os.ReadFile(path)
	if err != nil {
		return Status{}
	}
	var st Status
	if json.Unmarshal(b, &st) != nil {
		return Status{}
	}
	return st
}

func WriteStatus(path string, st Status) error {
	b, err := json.Marshal(st)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func ClearStatus(path string) {
	_ = os.Remove(path)
}

func Download(st store.Store, newVer, destDir string) error {
	host := st.GetString("storeIp", "127.0.0.1")
	port := st.GetInt("storePort", 21)
	if port <= 0 {
		port = 21
	}
	addr := host + ":" + itoa(port)
	dir := ftpRoot + "/version" + newVer
	log.Printf("upgrade ftp %s dir=%s", addr, dir)
	return ftpGet(addr, ftpUser, ftpPass, dir, destDir)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
