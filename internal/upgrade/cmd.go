package upgrade

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ParseApplyArgs(args []string) (ApplyOpts, bool) {
	found := false
	rest := make([]string, 0, len(args))
	for _, a := range args {
		if a == "--apply-update" {
			found = true
			continue
		}
		rest = append(rest, a)
	}
	if !found {
		return ApplyOpts{}, false
	}
	fs := flag.NewFlagSet("apply-update", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var opts ApplyOpts
	fs.IntVar(&opts.PID, "pid", 0, "")
	fs.StringVar(&opts.Target, "target", "", "")
	fs.StringVar(&opts.From, "from", "", "")
	fs.StringVar(&opts.Version, "version", "", "")
	fs.StringVar(&opts.LockDir, "lock-dir", "", "")
	if err := fs.Parse(rest); err != nil {
		return ApplyOpts{}, false
	}
	if strings.TrimSpace(opts.Target) == "" || strings.TrimSpace(opts.From) == "" {
		return ApplyOpts{}, false
	}
	return opts, true
}

// MaybeApply runs the updater if args contain --apply-update.
// Returns true if this process was the updater (caller should return from main).
func MaybeApply(args []string) bool {
	opts, ok := ParseApplyArgs(args)
	if !ok {
		return false
	}
	if opts.Target != "" {
		SetLogDir(filepath.Dir(opts.Target))
	}
	if err := Apply(opts); err != nil {
		logf("apply failed: %v", err)
		os.Exit(1)
	}
	return true
}
