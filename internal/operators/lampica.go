package operators

import (
	"strconv"
	"strings"
	"sync"

	"nts/vanguard/internal/print"
)

type lampicaOp struct {
	mu   sync.Mutex
	card string
}

func (o *lampicaOp) Owns(id string) bool {
	return strings.HasPrefix(id, "lampica")
}

func (o *lampicaOp) cardNumber() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.card
}

func (o *lampicaOp) setCard(n string) {
	o.mu.Lock()
	o.card = n
	o.mu.Unlock()
}

func (o *lampicaOp) OnCard(current, number string) (string, bool) {
	if current == "lampica.auto" {
		o.setCard(number)
		return "", true
	}
	if len(number) >= 3 && number[:3] == "567" {
		o.setCard(number)
		return "lampica.auto", true
	}
	return "", false
}

func (o *lampicaOp) View(id string, t Text) Screen {
	switch id {
	case "lampica.home":
		return menuScreen([]Item{
			navItem("lampica.auto", t("ui.swipe")),
			navItem("lampica.manual", t("ui.manual")),
			navItem("lampica.status", t("ui.points")),
			navItem("lampica.voucher", t("ui.coupon")),
			navItem("lampica.last", t("ui.copies")),
			navItem("lampica.charge", t("ui.storno")),
			reportItem("report.lamp_day", "rdL", "lampDay", t("ui.day")),
			reportItem("report.lamp_month", "mdL", "lampMonth", t("ui.month")),
		})
	case "lampica.auto":
		info := t("ui.swipe_card")
		if c := o.cardNumber(); c != "" {
			info = t("ui.card_n", c)
		}
		s := formScreen(Field{Name: "amt", Label: t("ui.amount"), Type: "text"})
		s.Info = info
		return s
	case "lampica.manual":
		return formScreen(
			Field{Name: "cn", Label: t("ui.serial"), Type: "text"},
			Field{Name: "amt", Label: t("ui.amount"), Type: "text"},
		)
	case "lampica.status":
		return formScreen(Field{Name: "cn", Label: t("ui.card_or_sn"), Type: "text"})
	case "lampica.voucher":
		return formScreen(
			Field{Name: "cn", Label: t("ui.card"), Type: "text"},
			Field{Name: "amt", Label: t("ui.amount"), Type: "text"},
		)
	case "lampica.last":
		return menuScreen([]Item{
			cmdItem("lampLast", "sL", t("ui.swipe_copy")),
			cmdItem("lampLast", "sLv", t("ui.coupon")),
		})
	case "lampica.charge":
		return formScreen(Field{Name: "stid", Label: t("ui.txid"), Type: "text"})
	default:
		return infoScreen(id)
	}
}

func (o *lampicaOp) Submit(id string, v map[string]string, rt Runtime) {
	switch id {
	case "lampica.auto":
		if o.cardNumber() == "" || strings.TrimSpace(v["amt"]) == "" {
			rt.Toast(rt.T("toast.swipe_amt"))
			return
		}
		o.send(rt, "rL", "crL", map[string]string{"cn": o.cardNumber(), "amt": v["amt"]}, rt.T("toast.tx_ok"), lampReceipt("Provlačenje"))
	case "lampica.manual":
		if len(strings.TrimSpace(v["cn"])) != 9 {
			rt.Toast(rt.T("toast.card_sn"))
			return
		}
		o.send(rt, "rL", "crL", map[string]string{"cn": v["cn"], "amt": v["amt"]}, rt.T("toast.tx_ok"), lampReceipt("Ručno"))
	case "lampica.status":
		rt.RunNet(rt.Socket().Put("com", "bL").Put("cn", v["cn"]), func(m map[string]string) {
			if m["err"] == "60" || m["err"] == "30" || m["err"] == "" {
				rt.Toast(rt.T("toast.balance", m["amt"]))
				rt.Print(m, lampReceipt("Stanje bodova"), "lamp", true)
				return
			}
			rt.Toast(rt.T(m["err"]))
		}, nil)
	case "lampica.voucher":
		o.send(rt, "rLv", "crLv", map[string]string{"cn": v["cn"], "amt": v["amt"]}, rt.T("toast.voucher_ok"), lampReceipt("Kupon"))
	case "lampica.charge":
		h := rt.Socket().Put("com", "cL").Put("stid", v["stid"]).Put("rc", strconv.Itoa(rt.NextRC()))
		rt.RunNet(h, func(m map[string]string) {
			if m["err"] == "50" || m["err"] == "30" {
				rt.Toast(rt.T("toast.storno_ok"))
				rt.Print(m, lampReceipt("Storno"), "lamp", true)
				rt.Home()
				return
			}
			rt.Toast(rt.T(m["err"]))
		}, nil)
	}
}

func (o *lampicaOp) Tap(_ string, act Action, rt Runtime) bool {
	if act.Kind == ActCmd && act.Target == "lampLast" {
		o.last(rt, act.Arg)
		return true
	}
	return false
}

func (o *lampicaOp) send(rt Runtime, cmd, check string, extra map[string]string, okMsg string, tmpl *print.Template) {
	rc := rt.NextRC()
	h := rt.Socket()
	for k, v := range extra {
		h.Put(k, v)
	}
	rt.DoAndWaitCheck(h, rc, cmd, check, func(m map[string]string) {
		if m["err"] == "30" {
			rt.Toast(okMsg)
			rt.PrintAll(m, tmpl, tmpl, "lamp")
			rt.Home()
			return
		}
		rt.Toast(rt.T(m["err"]))
		rt.Home()
	})
}

func (o *lampicaOp) last(rt Runtime, com string) {
	title := "Kopija provlačenja"
	if com == "sLv" {
		title = "Kopija kupona"
	}
	rt.RunNet(rt.Socket().Put("com", com), func(m map[string]string) {
		if m["err"] == "30" {
			rt.Toast(rt.T("toast.last_ok"))
			rt.Print(m, lampReceipt(title), "lamp", true)
			rt.Home()
			return
		}
		rt.Toast(rt.T(m["err"]))
	}, nil)
}

func lampReceipt(title string) *print.Template {
	b := print.NewBuilder()
	return b.Line().Text("LAMPICA").Line().Text(title).
		Line().Param("dts").
		Line().Param("mn").Text(",").Param("pa").
		Line().Text("Kartica: ").Param("cn").
		Line().Param("amt").Text(" KM").
		Line().Text("ID transakcije").Line().Param("tid").
		Line().Template()
}
