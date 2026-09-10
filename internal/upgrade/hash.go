package upgrade

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func FileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ParseHashFile accepts a raw 64-char hex digest or sha256sum format:
//
//	<hex>  filename
//	<hex> *filename
func ParseHashFile(b []byte) (string, error) {
	s := strings.TrimSpace(string(b))
	s = strings.TrimPrefix(s, "\ufeff")
	if s == "" {
		return "", fmt.Errorf("empty hash file")
	}
	line := s
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		line = strings.TrimSpace(s[:i])
	}
	hexPart := line
	if i := strings.IndexFunc(line, unicode.IsSpace); i >= 0 {
		hexPart = line[:i]
	}
	hexPart = strings.ToLower(strings.TrimSpace(hexPart))
	if len(hexPart) != sha256.Size*2 {
		return "", fmt.Errorf("invalid sha256 digest")
	}
	if _, err := hex.DecodeString(hexPart); err != nil {
		return "", fmt.Errorf("invalid sha256 digest")
	}
	return hexPart, nil
}

func VerifyFile(path, wantHex string) error {
	got, err := FileSHA256(path)
	if err != nil {
		return err
	}
	if !strings.EqualFold(got, wantHex) {
		return fmt.Errorf("sha256 mismatch")
	}
	return nil
}

func isHashName(name string) bool {
	n := strings.ToLower(name)
	switch n {
	case "sha256", "sha256.txt", "sha256sums":
		return true
	}
	return strings.HasSuffix(n, ".sha256")
}

func isSkipName(name string) bool {
	n := strings.ToLower(name)
	if isHashName(n) {
		return true
	}
	switch n {
	case "tools.jar", statusName:
		return true
	}
	return strings.HasSuffix(n, ".new") || strings.HasSuffix(n, ".bak") || strings.HasSuffix(n, ".bad")
}

// FindPayload picks the replacement binary and its sha256 file in dir.
func FindPayload(dir, exeName string) (payload, hashFile string, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", "", err
	}
	var files []string
	var hashes []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		path := filepath.Join(dir, name)
		if isHashName(name) {
			hashes = append(hashes, path)
			continue
		}
		if isSkipName(name) {
			continue
		}
		files = append(files, path)
	}
	if len(files) == 0 {
		return "", "", fmt.Errorf("no application file in %s", dir)
	}
	payload = pickPayload(files, exeName)
	hashFile = pickHash(hashes, payload)
	if hashFile == "" {
		return "", "", fmt.Errorf("missing sha256 file for %s", filepath.Base(payload))
	}
	return payload, hashFile, nil
}

func pickPayload(files []string, exeName string) string {
	want := strings.ToLower(exeName)
	for _, f := range files {
		if strings.EqualFold(filepath.Base(f), exeName) {
			return f
		}
	}
	fallbacks := []string{"npos.exe", "npos", "vanguard.exe", "vanguard", "n pos.exe", "n pos", "nts pos.exe", "nts pos"}
	for _, name := range fallbacks {
		for _, f := range files {
			if strings.ToLower(filepath.Base(f)) == name {
				return f
			}
		}
	}
	var exes []string
	for _, f := range files {
		if strings.EqualFold(filepath.Ext(f), ".exe") {
			exes = append(exes, f)
		}
	}
	if len(exes) == 1 {
		return exes[0]
	}
	if want != "" {
		for _, f := range files {
			base := strings.ToLower(filepath.Base(f))
			if strings.TrimSuffix(base, filepath.Ext(base)) == strings.TrimSuffix(want, filepath.Ext(want)) {
				return f
			}
		}
	}
	if len(files) == 1 {
		return files[0]
	}
	if len(exes) > 0 {
		return exes[0]
	}
	return files[0]
}

func pickHash(hashes []string, payload string) string {
	want := strings.ToLower(filepath.Base(payload) + ".sha256")
	for _, h := range hashes {
		if strings.ToLower(filepath.Base(h)) == want {
			return h
		}
	}
	for _, name := range []string{"sha256", "sha256.txt", "sha256sums"} {
		for _, h := range hashes {
			if strings.ToLower(filepath.Base(h)) == name {
				return h
			}
		}
	}
	if len(hashes) == 1 {
		return hashes[0]
	}
	return ""
}
