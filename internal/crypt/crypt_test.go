package crypt

import "testing"

func TestInvertRoundTrip(t *testing.T) {
	in := "tsN=123&pass=111&com=gV"
	got := Invert(Invert(in))
	if got != in {
		t.Fatalf("round trip: got %q want %q", got, in)
	}
}

func TestInvertXY(t *testing.T) {
	// Scala SimpleCrypt.main uses "xy"
	once := Invert("xy")
	if Invert(once) != "xy" {
		t.Fatalf("xy round trip failed: %q", once)
	}
}

func TestChecksum(t *testing.T) {
	line := []byte("a=1&b=2")
	var want int64
	for i, b := range line {
		want += int64(i+1) * int64(b&0xFF)
	}
	if got := Checksum(line); got != want {
		t.Fatalf("checksum: got %d want %d", got, want)
	}
}
