package print

import (
	"errors"
	"testing"
)

func TestWriteOSRequiresPrinter(t *testing.T) {
	for _, name := range []string{"", "  "} {
		j := &Job{Settings: Settings{PrinterName: name}, Content: "x"}
		if err := writeOS(j); !errors.Is(err, ErrNoPrinter) {
			t.Fatalf("printer %q: got %v", name, err)
		}
	}
}
