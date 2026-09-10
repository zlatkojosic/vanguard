package upgrade

import (
	"net"
	"testing"
)

func TestParsePASV(t *testing.T) {
	host, port, err := parsePASV("Entering Passive Mode (127,0,0,1,20,21).")
	if err != nil {
		t.Fatal(err)
	}
	if host != "127.0.0.1" || port != 20*256+21 {
		t.Fatalf("host=%s port=%d", host, port)
	}
}

func TestFormatPORT(t *testing.T) {
	got := formatPORT(net.IPv4(10, 1, 2, 3), 53759)
	if got != "10,1,2,3,209,255" {
		t.Fatalf("got %s", got)
	}
}

func TestIsNetTimeout(t *testing.T) {
	if !isNetTimeout(&timeoutErr{}) {
		t.Fatal("expected timeout")
	}
}

type timeoutErr struct{}

func (timeoutErr) Error() string   { return "i/o timeout" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return true }
