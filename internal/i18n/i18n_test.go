package i18n

import "testing"

func TestTranslateFallback(t *testing.T) {
	tr := New()
	if got := tr.T("chrome.login"); got != "Prijava" {
		t.Fatalf("sr login: %q", got)
	}
	if got := tr.T("chrome.back"); got != "Nazad" {
		t.Fatalf("sr back: %q", got)
	}
	tr.SetLang("en")
	if got := tr.T("chrome.login"); got != "Login" {
		t.Fatalf("en login: %q", got)
	}
	if got := tr.T("11"); got != "Wrong password!" {
		t.Fatalf("en err: %q", got)
	}
	tr.SetLang("hr")
	if got := tr.T("chrome.settings"); got != "Postavke" {
		t.Fatalf("hr settings: %q", got)
	}
	if got := tr.T("chrome.back"); got != "Natrag" {
		t.Fatalf("hr back: %q", got)
	}
	if got := tr.T("11"); got == "11" {
		t.Fatalf("hr should fall back for error 11")
	}
}

func TestFormat(t *testing.T) {
	tr := New()
	tr.SetLang("en")
	if got := tr.T("toast.language", "English"); got != "Language set to English" {
		t.Fatalf("format: %q", got)
	}
}

func TestNormalize(t *testing.T) {
	if Normalize("eng") != "en" || Normalize("srb") != "sr" {
		t.Fatalf("normalize")
	}
}
