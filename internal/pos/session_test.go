package pos

import (
	"testing"
	"time"

	"nts/vanguard/internal/operators"
	"nts/vanguard/internal/store"
)

func TestChromeAndOperatorScreens(t *testing.T) {
	s := New(store.NewMemory())
	if got := s.State().Kind; got != "tabs" {
		t.Fatalf("login kind %s", got)
	}

	s.Tap("go:theme")
	if got := s.State().Kind; got != "theme" {
		t.Fatalf("theme kind %s", got)
	}

	s.Tap("go:eronet.home")
	ui := s.State()
	if ui.Kind != "menu" || len(ui.Items) != 2 || ui.Items[0].ID != "go:eronet.topup" {
		t.Fatalf("eronet.home %+v", ui)
	}

	s.Tap("go:aircache.home")
	ui = s.State()
	if ui.Kind != "menu" || ui.Title != "Aircash" {
		t.Fatalf("aircache.home %+v", ui)
	}

	s.Tap("go:lampica.home")
	ui = s.State()
	if ui.Kind != "menu" || len(ui.Items) != 8 {
		t.Fatalf("lampica.home %+v", ui)
	}
	if ui.Labels["login"] != "Prijava" {
		t.Fatalf("labels %+v", ui.Labels)
	}

	s.mu.Lock()
	s.opCfg = operators.Overlay{Currency: "EUR", Amounts: map[string][]int{"xBon": {10, 20}}}
	s.stack = []screen{{ID: "xbon.voucher.recharge", Title: "XBon"}}
	s.mu.Unlock()
	ui = s.State()
	if ui.Kind != "menu" || len(ui.Items) != 2 || ui.Items[0].Title != "10 EUR" {
		t.Fatalf("overlay tiles %+v", ui)
	}

	s.Tap("lang:en")
	ui = s.State()
	if ui.Lang != "en" || ui.Labels["login"] != "Login" {
		t.Fatalf("english lang=%s labels=%+v", ui.Lang, ui.Labels)
	}
	if ui.Labels["back"] != "Back" {
		t.Fatalf("english back %q", ui.Labels["back"])
	}

	s.Tap("lang:sr")
	ui = s.State()
	if ui.Lang != "sr" || ui.Labels["back"] != "Nazad" {
		t.Fatalf("serbian back lang=%s labels=%+v", ui.Lang, ui.Labels)
	}

	s.Tap("go:system")
	ui = s.State()
	if ui.Kind != "form" {
		t.Fatalf("system kind %s", ui.Kind)
	}
	got := map[string]Field{}
	for _, f := range ui.Fields {
		got[f.Name] = f
	}
	for _, name := range []string{"serverIp", "serverPort", "transport", "scheme"} {
		if _, ok := got[name]; !ok {
			t.Fatalf("system missing %s: %+v", name, ui.Fields)
		}
	}
	if _, ok := got["userId"]; ok {
		t.Fatal("userId should not be on the form")
	}
	if got["transport"].Type != "select" || got["transport"].Value != "socket" {
		t.Fatalf("transport %+v", got["transport"])
	}

	s.Submit(map[string]string{
		"serverIp": "127.0.0.1", "serverPort": "8080", "transport": "rest",
		"scheme": "http",
	})
	if s.Pref.GetString("transport", "") != "rest" {
		t.Fatalf("save rest transport=%s", s.Pref.GetString("transport", ""))
	}
	if !s.useREST() {
		t.Fatal("expected rest after save")
	}
}

func TestUpgradeWaitsForButton(t *testing.T) {
	if skipSelfUpdate() {
		t.Skip("self-update is disabled in wails dev")
	}
	s := New(store.NewMemory())
	s.Pref.PutString("appVer", "1.0")
	s.Pref.PutString("storeIp", "10.0.0.1")
	s.applySession(map[string]string{"printable": "rE", "servisi": "rE", "appVer": "16.0"})
	if s.Pref.GetString("newAppVer", "") != "16.0" {
		t.Fatalf("newAppVer %s", s.Pref.GetString("newAppVer", ""))
	}
	s.mu.Lock()
	up := s.upgrading
	s.mu.Unlock()
	if up {
		t.Fatal("should not download until Upgrade is tapped")
	}

	s.mu.Lock()
	s.stack = []screen{{ID: "home", Title: "Home"}}
	s.tab = "apps"
	s.mu.Unlock()
	s.offerUpgrade()
	ui := s.State()
	if ui.Kind != "menu" || ui.Info != s.T("ui.upgrade_available") {
		t.Fatalf("prompt %+v", ui)
	}
	if len(ui.Items) != 1 || ui.Items[0].ID != "cmd:upgrade" || ui.Items[0].Title != s.T("ui.upgrade_now") {
		t.Fatalf("upgrade button %+v", ui.Items)
	}

	done := make(chan struct{})
	s.Quit = func() { close(done) }
	s.installUpgrade = func(_ store.Store, ver, _ string) error {
		if ver != "16.0" {
			t.Errorf("ver %s", ver)
		}
		return nil
	}
	s.Tap("cmd:upgrade")
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for quit after Upgrade")
	}
	if got := s.State().Message; got != s.T("toast.upgrade_restart") {
		t.Fatalf("message after upgrade %q", got)
	}
}
