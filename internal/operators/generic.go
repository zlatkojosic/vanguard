package operators

import (
	"strconv"
	"strings"
	"time"
)

type genericOp struct{}

func (genericOp) Owns(id string) bool {
	for _, o := range Catalog() {
		if o.Activity == id && o.Kind == KindHome {
			return true
		}
	}
	f, _, ok := FlowByRoute(id)
	return ok && f != nil
}

func (genericOp) View(id string, t Text) Screen {
	return flowScreen(id, t)
}

func (genericOp) Submit(id string, form map[string]string, rt Runtime) {
	switch {
	case strings.HasSuffix(id, ".charge"):
		runCharge(id, form["stid"], rt)
	case strings.HasSuffix(id, ".recharge"):
		runTopup(id, form, rt)
	}
}

func (genericOp) Tap(id string, act Action, rt Runtime) bool {
	switch act.Kind {
	case ActLast:
		runLast(act.Target, rt)
		return true
	case ActBuy:
		runAmountBuy(id, act.Target, rt, nil)
		return true
	default:
		return false
	}
}

func flowScreen(id string, t Text) Screen {
	if t == nil {
		t = Keys
	}
	for _, op := range Catalog() {
		if op.Activity == id && op.Kind == KindHome {
			s := Screen{Kind: "menu"}
			if op.Topup != nil {
				s.Items = append(s.Items, navItem(op.ID+".topup", t("ui.topup")))
			}
			if op.Voucher != nil {
				s.Items = append(s.Items, navItem(op.ID+".voucher", t("ui.voucher")))
			}
			return s
		}
	}
	f, opID, ok := FlowByRoute(id)
	if !ok || f == nil {
		return infoScreen(id)
	}
	isRecharge := strings.HasSuffix(id, ".recharge") || strings.HasSuffix(id, ".vrecharge")
	isCharge := strings.HasSuffix(id, ".charge")
	if !isRecharge && !isCharge {
		return flowMenu(id, opID, f, t)
	}
	if isCharge {
		return formScreen(Field{Name: "stid", Label: t("ui.txid"), Type: "text", MaxLen: f.StornoMax, Numeric: numericIf(f.StornoMax > 0, "int")})
	}
	if opID == "binbin" || strings.HasPrefix(id, "binbin") {
		return binbinProducts()
	}
	if f.Amounts != nil {
		s := Screen{Kind: "menu"}
		for _, a := range f.Amounts {
			s.Items = append(s.Items, buyItem(strconv.Itoa(a), strconv.Itoa(a)+" "+DefaultCurrency))
		}
		return s
	}
	amtNumeric := "decimal"
	if f.IntAmount {
		amtNumeric = "int"
	}
	return formScreen(
		Field{Name: "phN", Label: t("ui.phone"), Type: "text", MaxLen: f.PhoneMax, Numeric: numericIf(f.PhoneMax > 0, "int")},
		Field{Name: "amt", Label: t("ui.amount"), Type: "text", Numeric: amtNumeric},
	)
}

func numericIf(ok bool, kind string) string {
	if ok {
		return kind
	}
	return ""
}

// validDigits reports whether s is non-empty, digits only and at most max long (max <= 0 means no limit).
func validDigits(s string, max int) bool {
	if s == "" || (max > 0 && len(s) > max) {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func flowMenu(id, opID string, f *Flow, t Text) Screen {
	kind := "topup"
	if f.Amounts != nil {
		kind = "voucher"
	}
	s := Screen{Kind: "menu", Items: []Item{navItem(id+".recharge", t("ui.recharge"))}}
	if kind == "topup" && f.HasCharge {
		s.Items = append(s.Items, navItem(id+".charge", t("ui.storno")))
	}
	s.Items = append(s.Items,
		lastItem(opID+":"+kind, t("ui.copy")),
		reportItem("report.sum", "rsA", "sum", t("ui.summary")),
		reportItem("report.detail", f.Report, "detail", t("ui.detail")),
	)
	if kind == "voucher" && f.HasCharge {
		chargeID := id + ".charge"
		if strings.HasPrefix(id, "xbon.sms") {
			chargeID = "xbon.sms.charge"
		}
		s.Items = append(s.Items, navItem(chargeID, t("ui.storno")))
	}
	return s
}

func runTopup(id string, v map[string]string, rt Runtime) {
	f, _, ok := FlowByRoute(id)
	if !ok || f == nil {
		return
	}
	if strings.TrimSpace(v["phN"]) == "" || strings.TrimSpace(v["amt"]) == "" {
		rt.Toast(rt.T("toast.enter_phone_amt"))
		return
	}
	phone, amt := strings.TrimSpace(v["phN"]), strings.TrimSpace(v["amt"])
	if f.PhoneMax > 0 && !validDigits(phone, f.PhoneMax) {
		rt.Toast(rt.T("toast.invalid_phone", f.PhoneMax))
		return
	}
	if f.IntAmount && !validDigits(amt, 0) {
		rt.Toast(rt.T("toast.invalid_amount"))
		return
	}
	rt.Ask(confirmText(f, rt), func() {
		if !rt.AllowBuy() {
			return
		}
		rt.SetAllowBuy(false)
		rc := rt.NextRC()
		h := rt.Socket().Put("amt", amt).Put("phN", phone)
		rt.DoAndWaitCheck(h, rc, f.Command, f.Check, func(m map[string]string) {
			if m["err"] == "30" {
				rt.Toast(rt.T("toast.topup_ok"))
				rt.PrintAll(m, f.Full(), f.Short(), f.Service)
				rt.Home()
				return
			}
			rt.Toast(rt.T(m["err"]))
			rt.Home()
		})
	})
}

func runAmountBuy(screenID, amt string, rt Runtime, onOK func(map[string]string)) {
	f, opID, ok := FlowByRoute(screenID)
	if !ok || f == nil || !rt.AllowBuy() {
		return
	}
	rt.Ask(confirmText(f, rt), func() {
		rt.SetAllowBuy(false)
		rc := rt.NextRC()
		h := rt.Socket().Put("amt", amt)
		rt.DoAndWaitCheck(h, rc, f.Command, f.Check, func(m map[string]string) {
			if m["err"] == "30" {
				if onOK != nil {
					onOK(m)
				} else if opID == "xbon1" || strings.HasPrefix(screenID, "xbon.sms") {
					saveXBon(append(loadXBon(), xbonTx{ID: m["tid"], Amount: amt, Inserted: time.Now()}))
				}
				rt.Toast(rt.T("toast.topup_ok"))
				rt.PrintAll(m, f.Full(), f.Short(), f.Service)
				rt.Home()
				return
			}
			rt.SetAllowBuy(true)
			rt.Toast(rt.T(m["err"]))
		})
	})
}

func runCodeBuy(screenID, code string, rt Runtime) {
	f, _, ok := FlowByRoute(screenID)
	if !ok || f == nil || !rt.AllowBuy() {
		return
	}
	rt.Ask(confirmText(f, rt), func() {
		rt.SetAllowBuy(false)
		rc := rt.NextRC()
		h := rt.Socket().Put("code", code)
		rt.DoAndWaitCheck(h, rc, f.Command, f.Check, func(m map[string]string) {
			if m["err"] == "30" {
				rt.Toast(rt.T("toast.topup_ok"))
				rt.PrintAll(m, f.Full(), f.Short(), f.Service)
				rt.Home()
				return
			}
			rt.SetAllowBuy(true)
			rt.Toast(rt.T(m["err"]))
		})
	})
}

func runCharge(id, stid string, rt Runtime) {
	if strings.TrimSpace(stid) == "" {
		return
	}
	f, _, ok := FlowByRoute(id)
	if !ok || f == nil {
		return
	}
	stid = strings.TrimSpace(stid)
	if f.StornoMax > 0 && !validDigits(stid, f.StornoMax) {
		rt.Toast(rt.T("toast.invalid_stid", f.StornoMax))
		return
	}
	h := rt.Socket().Put("com", f.Charge).Put("stid", stid).Put("rc", strconv.Itoa(rt.NextRC()))
	rt.RunNet(h, func(m map[string]string) {
		want := f.SuccessCharge
		if want == "" {
			want = "50"
		}
		if m["err"] == want {
			rt.Toast(rt.T("toast.storno_ok"))
			rt.Print(m, f.ChargeT(), f.Service, true)
			rt.Home()
			return
		}
		rt.Toast(rt.T(m["err"]))
	}, nil)
}

func runLast(spec string, rt Runtime) {
	var f *Flow
	if spec == "sale" {
		f, _, _ = FlowByRoute("sale.home")
	} else {
		parts := strings.Split(spec, ":")
		kind := "topup"
		opID := spec
		if len(parts) == 2 {
			opID, kind = parts[0], parts[1]
		}
		f, _, _ = FlowByRoute(opID + "." + kind)
	}
	if f == nil {
		return
	}
	rt.RunNet(rt.Socket().Put("com", f.Last), func(m map[string]string) {
		if m["err"] == "30" {
			rt.Toast(rt.T("toast.last_ok"))
			rt.Print(m, f.LastT(), f.Service, true)
			rt.Home()
			return
		}
		rt.Toast(rt.T(m["err"]))
	}, nil)
}
