package operators

import (
	"strings"
)

type saleOp struct{}

func (saleOp) Owns(id string) bool {
	return strings.HasPrefix(id, "sale")
}

func (saleOp) View(id string, t Text) Screen {
	switch id {
	case "sale.home":
		cmd := "rdPSell"
		if f, _, ok := FlowByRoute(id); ok && f != nil {
			cmd = f.Report
		}
		return menuScreen([]Item{
			navItem("sale.sell", t("ui.sale")),
			navItem("sale.charge", t("ui.storno")),
			lastItem("sale", t("ui.copy")),
			reportItem("report.sum", "rsA", "sum", t("ui.summary")),
			reportItem("report.detail", cmd, "detail", t("ui.detail")),
		})
	case "sale.sell":
		return formScreen(
			Field{Name: "code", Label: t("ui.product_code"), Type: "text"},
			Field{Name: "qu", Label: t("ui.qty"), Type: "text"},
		)
	case "sale.charge":
		return formScreen(Field{Name: "stid", Label: t("ui.txid"), Type: "text"})
	default:
		return infoScreen(id)
	}
}

func (saleOp) Submit(id string, v map[string]string, rt Runtime) {
	switch id {
	case "sale.sell":
		saleSell(rt, v)
	case "sale.charge":
		runCharge(id, v["stid"], rt)
	}
}

func (saleOp) Tap(_ string, act Action, rt Runtime) bool {
	if act.Kind == ActLast {
		runLast(act.Target, rt)
		return true
	}
	return false
}

func saleSell(rt Runtime, v map[string]string) {
	c := strings.TrimSpace(v["code"])
	if len(c) != 2 && len(c) != 4 && len(c) != 20 {
		rt.Toast(rt.T("toast.sale_code"))
		return
	}
	f, _, ok := FlowByRoute("sale.sell")
	if !ok || f == nil {
		return
	}
	qty := v["qu"]
	rt.Ask(confirmText(f, rt), func() {
		rt.SetAllowBuy(false)
		rc := rt.NextRC()
		h := rt.Socket().Put("qu", qty).Put("code", c)
		rt.DoAndWaitCheck(h, rc, f.Command, f.Check, func(m map[string]string) {
			if m["err"] == "30" {
				rt.Toast(rt.T("toast.sale_ok"))
				rt.PrintAll(m, f.Full(), f.Short(), f.Service)
				rt.Home()
				return
			}
			rt.Toast(rt.T(m["err"]))
			rt.Home()
		})
	})
}
