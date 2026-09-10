package upgrade

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseHashFile(t *testing.T) {
	sum := strings.Repeat("ab", 32)
	cases := []struct {
		in   string
		ok   bool
		want string
	}{
		{sum, true, sum},
		{sum + "  NPOS.exe\n", true, sum},
		{sum + " *vanguard.exe\n", true, sum},
		{"\ufeff" + strings.ToUpper(sum) + "\n", true, sum},
		{"not-a-hash", false, ""},
		{"", false, ""},
	}
	for _, c := range cases {
		got, err := ParseHashFile([]byte(c.in))
		if c.ok {
			if err != nil {
				t.Fatalf("parse %q: %v", c.in, err)
			}
			if got != c.want {
				t.Fatalf("got %s want %s", got, c.want)
			}
		} else if err == nil {
			t.Fatalf("expected error for %q", c.in)
		}
	}
}

func TestFindPayloadAndVerify(t *testing.T) {
	dir := t.TempDir()
	payload := filepath.Join(dir, "NPOS.exe")
	body := []byte("fake-exe-body")
	if err := os.WriteFile(payload, body, 0o644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(body)
	hexSum := hex.EncodeToString(sum[:])
	if err := os.WriteFile(payload+".sha256", []byte(hexSum+"  NPOS.exe\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, hash, err := FindPayload(dir, "NPOS.exe")
	if err != nil {
		t.Fatal(err)
	}
	if got != payload {
		t.Fatalf("payload %s", got)
	}
	parsed, err := ParseHashFile(mustRead(t, hash))
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyFile(got, parsed); err != nil {
		t.Fatal(err)
	}
}

func TestFindPayloadRequiresHash(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "vanguard.exe"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := FindPayload(dir, "vanguard.exe"); err == nil {
		t.Fatal("expected missing hash error")
	}
}

func TestStageAndSwap(t *testing.T) {
	root := t.TempDir()
	exe := filepath.Join(root, "vanguard")
	if err := os.WriteFile(exe, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	p, err := ForExecutable(exe)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(p.Staging, 0o755); err != nil {
		t.Fatal(err)
	}
	body := []byte("new-binary")
	if err := os.WriteFile(filepath.Join(p.Staging, "vanguard"), body, 0o755); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(body)
	if err := os.WriteFile(filepath.Join(p.Staging, "vanguard.sha256"), []byte(hex.EncodeToString(sum[:])), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Stage(p, "1.1"); err != nil {
		t.Fatal(err)
	}
	st := ReadStatus(p.Status)
	if st.State != StateReady || st.Version != "1.1" {
		t.Fatalf("status %+v", st)
	}
	if _, err := os.Stat(p.New); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p.Staging); !os.IsNotExist(err) {
		t.Fatalf("staging should be removed: %v", err)
	}
	if err := swap(p); err != nil {
		t.Fatal(err)
	}
	if got := mustRead(t, p.Exe); string(got) != "new-binary" {
		t.Fatalf("exe %s", got)
	}
	if got := mustRead(t, p.Bak); string(got) != "old" {
		t.Fatalf("bak %s", got)
	}
	if err := restore(p); err != nil {
		t.Fatal(err)
	}
	if got := mustRead(t, p.Exe); string(got) != "old" {
		t.Fatalf("restored %s", got)
	}
}

func TestParseApplyArgs(t *testing.T) {
	opts, ok := ParseApplyArgs([]string{"--apply-update", "--pid", "42", "--target", "/a/vanguard", "--from", "/a/vanguard.new", "--version", "1.1", "--lock-dir", "/cfg"})
	if !ok {
		t.Fatal("expected apply args")
	}
	if opts.PID != 42 || opts.Version != "1.1" || opts.Target != "/a/vanguard" {
		t.Fatalf("%+v", opts)
	}
	if _, ok := ParseApplyArgs([]string{"--other"}); ok {
		t.Fatal("should ignore normal args")
	}
}

func TestVerifyMismatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f")
	if err := os.WriteFile(path, []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := VerifyFile(path, strings.Repeat("00", 32)); err == nil {
		t.Fatal("expected mismatch")
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
