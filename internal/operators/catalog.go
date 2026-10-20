package operators

import (
	"strings"

	"nts/vanguard/internal/print"
)

type Kind int

const (
	KindHome Kind = iota
	KindTopup
	KindVoucher
	KindLampica
	KindAircash
	KindSale
	KindBinBin
	KindXBonSMS
)

type Flow struct {
	Title         string
	Service       string
	Command       string
	Check         string
	Last          string
	Report        string
	Charge        string
	Amounts       []int
	Decimal       int
	Min, Max      float64
	Dialog        string
	OperatorName  string
	HasCharge     bool
	SuccessCharge string
	Receipt       string
	PhoneMax      int
	StornoMax     int
	IntAmount     bool
}

func (f Flow) Full() *print.Template {
	kind := "topup"
	if f.Amounts != nil {
		kind = "voucher"
	}
	return print.Receipt(f.Receipt, kind)
}

func (f Flow) Short() *print.Template { return f.Full() }
func (f Flow) LastT() *print.Template { return print.Receipt(f.Receipt, "last") }
func (f Flow) ChargeT() *print.Template {
	return print.Receipt(f.Receipt, "charge")
}

type Operator struct {
	ID, Title, Activity string
	Tags                []string
	Kind                Kind
	Topup, Voucher      *Flow
}

func Catalog() []Operator {
	return []Operator{
		{ID: "eronet", Title: "Eronet", Activity: "eronet.home", Tags: []string{"ero", "erot"}, Kind: KindHome,
			Topup:   &Flow{Title: "Eronet TopUp", Service: "erot", Command: "rEt", Check: "crEt", Last: "sEt", Report: "rdEt", Charge: "cEt", Decimal: 0, Min: 1, Max: 50, PhoneMax: 10, StornoMax: 14, IntAmount: true, OperatorName: "ERONET Topup", HasCharge: true, Receipt: "Eronet prepaid dopuna:"},
			Voucher: &Flow{Title: "Eronet Voucher", Service: "ero", Command: "rE", Check: "crE", Last: "sEr", Report: "rdEr", Amounts: []int{2, 5, 10, 20, 50}, OperatorName: "ERONET Voucher", Receipt: "Eronet voucher:"},
		},
		{ID: "mtel", Title: "Mtel", Activity: "mtel.topup", Tags: []string{"mtel"}, Kind: KindTopup,
			Topup: &Flow{Title: "Mtel TopUp", Service: "mtel", Command: "rM", Check: "crM", Last: "sMt", Report: "rdMt", Charge: "cMt", Decimal: 2, Min: 1, Max: 50, OperatorName: "MTEL Topup", HasCharge: true, Receipt: "Mtel prepaid dopuna:"},
		},
		{ID: "bht", Title: "Bht", Activity: "bht.home", Tags: []string{"bhtv", "bhttu"}, Kind: KindHome,
			Topup:   &Flow{Title: "Bht TopUp", Service: "bhttu", Command: "rB", Check: "crB", Last: "sBh", Report: "rdBh", Charge: "cBh", Decimal: 1, Min: 1, Max: 50, OperatorName: "BHT Topup", HasCharge: true, Receipt: "BHT prepaid dopuna:"},
			Voucher: &Flow{Title: "Bht Voucher", Service: "bhtv", Command: "rBv", Check: "crBv", Last: "bvs", Report: "rdBv", Amounts: []int{5, 10, 20, 50}, OperatorName: "BHT Voucher", Receipt: "BHT voucher:"},
		},
		{ID: "haloo", Title: "Haloo", Activity: "haloo.home", Tags: []string{"haloo", "halt"}, Kind: KindHome,
			Topup:   &Flow{Title: "Haloo TopUp", Service: "halt", Command: "rHt", Check: "crHt", Last: "sHt", Report: "rdHt", Charge: "cHt", Decimal: 0, Min: 1, Max: 50, OperatorName: "HALOO Topup", HasCharge: true, Receipt: "Haloo prepaid dopuna:"},
			Voucher: &Flow{Title: "Haloo Voucher", Service: "haloo", Command: "rHv", Check: "crHv", Last: "sHv", Report: "rdHv", Amounts: []int{1, 2, 5, 10, 20, 50}, OperatorName: "HALOO Voucher", Receipt: "Haloo voucher:"},
		},
		{ID: "novotel", Title: "Novotel", Activity: "novotel.topup", Tags: []string{"rNvt"}, Kind: KindTopup,
			Topup: &Flow{Title: "Novotel TopUp", Service: "rNvt", Command: "rNvt", Check: "crNvt", Last: "lNvt", Report: "rdNvt", Charge: "cNvt", Decimal: 2, Min: 1, Max: 50, OperatorName: "Novotel Topup", HasCharge: true, Receipt: "Novotel prepaid dopuna:"},
		},
		{ID: "izi", Title: "Izi Mobil", Activity: "izi.topup", Tags: []string{"izi"}, Kind: KindTopup,
			Topup: &Flow{Title: "Izi TopUp", Service: "izi", Command: "rI", Check: "crI", Last: "sI", Report: "rdI", Charge: "cI", Decimal: 2, Min: 1, Max: 50, OperatorName: "IZI Topup", HasCharge: true, Receipt: "Izi prepaid dopuna:"},
		},
		{ID: "happy", Title: "Happy", Activity: "happy.home", Tags: []string{"rY", "rYv"}, Kind: KindHome,
			Topup:   &Flow{Title: "Happy TopUp", Service: "rY", Command: "rY", Check: "crY", Last: "sYh", Report: "rdYh", Charge: "cYh", Decimal: 0, Min: 1, Max: 50, OperatorName: "HAPPY Topup", HasCharge: true, Receipt: "Happy prepaid dopuna:"},
			Voucher: &Flow{Title: "Happy Voucher", Service: "rYv", Command: "rYv", Check: "crYv", Last: "Yvs", Report: "rdvs", Amounts: []int{2, 5, 10, 20, 50}, OperatorName: "HAPPY Voucher", Receipt: "Happy voucher:"},
		},
		{ID: "xbon", Title: "XBon", Activity: "xbon.voucher", Tags: []string{"xBon"}, Kind: KindVoucher,
			Voucher: &Flow{Title: "XBon Voucher", Service: "xBon", Command: "sNv", Check: "crNv", Last: "lNv", Report: "rdNv", Charge: "cXb", Amounts: []int{5, 10, 20, 50, 100}, OperatorName: "XBon", HasCharge: true, Receipt: "XBon voucher:"},
		},
		{ID: "xbon1", Title: "XBon", Activity: "xbon.sms", Tags: []string{"xBon1"}, Kind: KindXBonSMS,
			Voucher: &Flow{Title: "XBon SMS", Service: "xBon", Command: "sNv", Check: "crNv", Last: "lNv", Report: "rdNv", Charge: "cXb", Amounts: []int{5, 10, 20, 50, 100}, OperatorName: "XBon", HasCharge: true, Receipt: "XBon voucher:"},
		},
		{ID: "abon", Title: "ABon", Activity: "abon.voucher", Tags: []string{"aBon"}, Kind: KindVoucher,
			Voucher: &Flow{Title: "ABon Voucher", Service: "aBon", Command: "startRechargeABon", Check: "startCheckABon", Last: "lastABon", Report: "detailReportABon", Amounts: []int{5, 10, 25, 50, 100, 200}, OperatorName: "ABon", Receipt: "ABon voucher:"},
		},
		{ID: "kliker", Title: "Kliker", Activity: "kliker.voucher", Tags: []string{"spiner"}, Kind: KindVoucher,
			Voucher: &Flow{Title: "Kliker Voucher", Service: "spiner", Command: "rechargeSpiner", Check: "checkSpiner", Last: "lastSpiner", Report: "rdSpiner", Charge: "chargeSpiner", Amounts: []int{5, 10, 20, 50, 100}, OperatorName: "Kliker", HasCharge: true, Receipt: "Kliker voucher:"},
		},
		{ID: "abc", Title: "ABC", Activity: "abc.voucher", Tags: []string{"rAbc"}, Kind: KindVoucher,
			Voucher: &Flow{Title: "ABC Voucher", Service: "rAbc", Command: "rAbc", Check: "scAbc", Last: "lAbc", Report: "rdAbc", Charge: "cAbc", Amounts: []int{5, 10, 20, 50, 100}, OperatorName: "ABC", HasCharge: true, Receipt: "ABC voucher:"},
		},
		{ID: "mozz", Title: "Mozzart", Activity: "mozz.voucher", Tags: []string{"newVMozz"}, Kind: KindVoucher,
			Voucher: &Flow{Title: "Mozzart Voucher", Service: "newVMozz", Command: "newVMozz", Check: "checkVMozz", Last: "lastVMozz", Report: "rdMozz", Charge: "chargeVMozz", Amounts: []int{5, 10, 20, 50, 100}, OperatorName: "Mozzart", HasCharge: true, Receipt: "Mozzart voucher:"},
		},
		{ID: "binbin", Title: "BinBin", Activity: "binbin.voucher", Tags: []string{"BnBn"}, Kind: KindBinBin,
			Voucher: &Flow{Title: "BinBin Voucher", Service: "BnBn", Command: "rBnBn", Check: "crBnBn", Last: "lBnBn", Report: "rdBnBn", OperatorName: "BinBin", Receipt: "BinBin voucher:"},
		},
		{ID: "sale", Title: "Prodaja", Activity: "sale.home", Tags: []string{"psell"}, Kind: KindSale,
			Topup: &Flow{Title: "Prodaja", Service: "psell", Command: "psellSell", Check: "psellCheck", Last: "psellLast", Report: "rdPSell", Charge: "psellStorno", Decimal: 2, Min: 1, Max: 100, Dialog: "dialog.sale", OperatorName: "Prodaja", HasCharge: true, SuccessCharge: "30", Receipt: "Prodaja proizvoda:"},
		},
		{ID: "lampica", Title: "Lampica", Activity: "lampica.home", Tags: []string{"lamp"}, Kind: KindLampica},
		{ID: "aircache", Title: "Aircash", Activity: "aircache.home", Tags: []string{"aircache"}, Kind: KindAircash},
	}
}

type Meta struct {
	Title, Activity string
	Tags            []string
	Always          bool
}

func Metas() []Meta {
	var out []Meta
	for _, o := range Catalog() {
		out = append(out, Meta{Title: o.Title, Activity: o.Activity, Tags: o.Tags})
	}
	return out
}

func FlowByRoute(id string) (*Flow, string, bool) {
	for _, o := range Catalog() {
		if id == o.ID+".topup" || id == o.ID+".topup.recharge" || id == o.ID+".topup.charge" {
			return o.Topup, o.ID, true
		}
		if id == o.ID+".voucher" || id == o.ID+".voucher.recharge" || id == o.ID+".voucher.charge" {
			return o.Voucher, o.ID, true
		}
		if o.Kind == KindTopup && (id == o.Activity || id == o.Activity+".recharge" || id == o.Activity+".charge") {
			return o.Topup, o.ID, true
		}
		if o.Kind == KindVoucher && (id == o.Activity || id == o.Activity+".recharge" || id == o.Activity+".charge") {
			return o.Voucher, o.ID, true
		}
		if o.Kind == KindXBonSMS && strings.HasPrefix(id, "xbon.sms") {
			return o.Voucher, o.ID, true
		}
		if o.Kind == KindBinBin && strings.HasPrefix(id, "binbin") {
			return o.Voucher, o.ID, true
		}
		if o.Kind == KindSale && strings.HasPrefix(id, "sale") {
			return o.Topup, o.ID, true
		}
	}
	return nil, "", false
}

func Title(id string, t Text) string {
	if t == nil {
		t = Keys
	}
	switch id {
	case "lampica.home":
		return "Lampica"
	case "lampica.auto":
		return t("title.lampica.auto")
	case "lampica.manual":
		return t("title.lampica.manual")
	case "lampica.status":
		return t("title.lampica.status")
	case "lampica.voucher":
		return t("title.lampica.voucher")
	case "lampica.last":
		return t("title.lampica.last")
	case "lampica.charge":
		return t("title.lampica.charge")
	case "aircache.home":
		return "Aircash"
	case "aircache.code":
		return t("title.aircache.code")
	case "aircache.storno":
		return t("title.aircache.storno")
	case "sale.home":
		return t("ui.sale")
	case "sale.sell":
		return t("title.sale.sell")
	case "xbon.sms.charge":
		return t("title.xbon.sms.charge")
	}
	for _, op := range Catalog() {
		if op.Activity == id {
			return op.Title
		}
		if op.Topup != nil {
			if id == op.ID+".topup" {
				return op.Topup.Title
			}
			if id == op.ID+".topup.recharge" || id == op.Activity+".recharge" {
				return op.Topup.Title + " " + t("ui.recharge")
			}
			if id == op.ID+".topup.charge" || strings.HasSuffix(id, ".charge") && strings.Contains(id, op.ID) {
				return op.Topup.Title + " " + t("ui.storno")
			}
		}
		if op.Voucher != nil {
			if id == op.ID+".voucher" {
				return op.Voucher.Title
			}
			if id == op.ID+".voucher.recharge" {
				return op.Voucher.Title + " " + t("ui.recharge")
			}
			if id == op.ID+".voucher.charge" {
				return op.Voucher.Title + " " + t("ui.storno")
			}
		}
	}
	return id
}
