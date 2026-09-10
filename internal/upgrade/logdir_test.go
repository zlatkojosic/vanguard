package upgrade

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenAppLogWritesFile(t *testing.T) {
	dir := t.TempDir()
	SetLogDir(dir)
	t.Cleanup(func() { SetLogDir("") })
	c := OpenAppLog()
	if c == nil {
		t.Fatal("OpenAppLog")
	}
	log.Println("hello from test")
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, appLogName))
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, "log file") || !strings.Contains(s, "hello from test") {
		t.Fatalf("log contents %q", s)
	}
}
