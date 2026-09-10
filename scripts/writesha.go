package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: writesha <binary>")
		os.Exit(1)
	}
	path := os.Args[1]
	st, err := os.Stat(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if st.IsDir() {
		fmt.Fprintf(os.Stderr, "skip sha256 for directory %s\n", path)
		return
	}
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		f.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	f.Close()
	sum := hex.EncodeToString(h.Sum(nil))
	name := filepath.Base(path)
	out := path + ".sha256"
	line := sum + "  " + name + "\n"
	if err := os.WriteFile(out, []byte(line), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(line)
}
