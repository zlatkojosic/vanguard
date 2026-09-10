package print

import "testing"

func TestRenderContainsBrand(t *testing.T) {
	tmpl := Receipt("Eronet prepaid dopuna:", "topup")
	out := Render(tmpl, Params{"dts": "01.01.2026", "mn": "Shop", "pa": "Ulica 1", "amt": "10", "tid": "99", "phN": "061123"}, 30)
	if out == "" {
		t.Fatal("empty render")
	}
	if !contains(out, "Eronet") || !contains(out, "10") {
		t.Fatalf("render: %s", out)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || search(s, sub))
}

func search(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
