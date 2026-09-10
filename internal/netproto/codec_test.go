package netproto

import "testing"

func TestMapRoundTrip(t *testing.T) {
	kvs := []KV{{"tsN", "1"}, {"pass", "111"}, {"com", "gV"}}
	line := MapToLine(kvs)
	if line != "tsN=1&pass=111&com=gV" {
		t.Fatalf("line: %q", line)
	}
	m := LineToMap(line)
	if m["tsN"] != "1" || m["pass"] != "111" || m["com"] != "gV" {
		t.Fatalf("map: %#v", m)
	}
}

func TestEmptyValue(t *testing.T) {
	m := LineToMap("err=&foo=bar")
	if m["err"] != "" {
		t.Fatalf("empty value: %q", m["err"])
	}
	if m["foo"] != "bar" {
		t.Fatalf("foo: %q", m["foo"])
	}
}

func TestEncodeDecode(t *testing.T) {
	line := "tsN=abc&com=gV"
	enc := EncodeLine(line)
	m := DecodeLine(enc)
	if m["tsN"] != "abc" || m["com"] != "gV" {
		t.Fatalf("decoded %#v", m)
	}
}
