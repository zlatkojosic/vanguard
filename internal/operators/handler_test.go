package operators

import "testing"

func TestDecodeEncode(t *testing.T) {
	cases := []Action{
		{Kind: ActNav, Target: "aircache.code"},
		{Kind: ActTheme, Target: "theme3.css"},
		{Kind: ActLang, Target: "en"},
		{Kind: ActReport, Target: "report.sum", Arg: "rsA", Extra: "sum"},
		{Kind: ActLast, Target: "eronet:topup"},
		{Kind: ActBuy, Target: "10"},
		{Kind: ActCode, Target: "100019"},
		{Kind: ActCmd, Target: "pending"},
		{Kind: ActCmd, Target: "copy"},
		{Kind: ActCmd, Target: "lampLast", Arg: "sL"},
		{Kind: ActCmd, Target: "storno", Arg: "tid-1"},
		{Kind: ActCmd, Target: "reportShow"},
		{Kind: ActCmd, Target: "reportPrint"},
	}
	for _, want := range cases {
		got := Decode(Encode(want))
		if got != want {
			t.Errorf("roundtrip %v: got %+v", Encode(want), got)
		}
	}
}

func TestDecodeLegacy(t *testing.T) {
	if got := Decode("acPending"); got.Kind != ActCmd || got.Target != "pending" {
		t.Fatalf("acPending: %+v", got)
	}
	if got := Decode("go:lampica.home"); got.Kind != ActNav || got.Target != "lampica.home" {
		t.Fatalf("go: %+v", got)
	}
}

func TestOperatorViews(t *testing.T) {
	home := View("eronet.home", Keys)
	if home.Kind != "menu" || len(home.Items) != 2 {
		t.Fatalf("eronet.home: %+v", home)
	}
	if home.Items[0].ID != "go:eronet.topup" || home.Items[1].ID != "go:eronet.voucher" {
		t.Fatalf("eronet.home items: %+v", home.Items)
	}

	mtel := View("mtel.topup", Keys)
	if mtel.Kind != "menu" || mtel.Items[0].ID != "go:mtel.topup.recharge" {
		t.Fatalf("mtel.topup: %+v", mtel)
	}

	voucher := View("xbon.voucher.recharge", Keys)
	if voucher.Kind != "menu" || len(voucher.Items) != 5 {
		t.Fatalf("xbon amounts: %+v", voucher)
	}
	if voucher.Items[0].ID != "amt:5" {
		t.Fatalf("first amount: %s", voucher.Items[0].ID)
	}

	over := ApplyOverlay("xbon.voucher.recharge", voucher, Overlay{
		Currency: "EUR",
		Amounts:  map[string][]int{"xBon": {10, 20}},
	})
	if len(over.Items) != 2 || over.Items[0].ID != "amt:10" || over.Items[0].Title != "10 EUR" {
		t.Fatalf("overlay: %+v", over)
	}

	lamp := View("lampica.home", Keys)
	if lamp.Kind != "menu" || len(lamp.Items) != 8 {
		t.Fatalf("lampica.home: %+v", lamp)
	}

	air := View("aircache.home", Keys)
	if air.Kind != "menu" {
		t.Fatalf("aircache.home kind %s", air.Kind)
	}
	found := false
	for _, it := range air.Items {
		if it.ID == "go:aircache.code" {
			found = true
		}
	}
	if !found {
		t.Fatalf("aircache.home missing code tile: %+v", air.Items)
	}

	sale := View("sale.home", Keys)
	if sale.Items[0].ID != "go:sale.sell" {
		t.Fatalf("sale.home: %+v", sale.Items)
	}

	bin := View("binbin.voucher.recharge", Keys)
	if bin.Columns != 2 || bin.Items[0].ID != "bin:100019" {
		t.Fatalf("binbin products: %+v", bin)
	}

	if Title("lampica.manual", Keys) != "title.lampica.manual" {
		t.Fatalf("title: %q", Title("lampica.manual", Keys))
	}
}

func TestLookupOrder(t *testing.T) {
	if _, ok := Lookup("sale.home").(saleOp); !ok {
		t.Fatalf("sale.home handler %T", Lookup("sale.home"))
	}
	if _, ok := Lookup("xbon.sms.charge").(xbonSMSOp); !ok {
		t.Fatalf("xbon.sms.charge handler %T", Lookup("xbon.sms.charge"))
	}
	if _, ok := Lookup("eronet.home").(genericOp); !ok {
		t.Fatalf("eronet.home handler %T", Lookup("eronet.home"))
	}
}
