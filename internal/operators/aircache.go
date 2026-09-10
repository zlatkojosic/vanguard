package operators

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"nts/vanguard/internal/print"
)

const aircacheFile = "abcdt.json"

type pendingTx struct {
	Code, Barcode string
}

type aircacheOp struct{}

func (aircacheOp) Owns(id string) bool {
	return strings.HasPrefix(id, "aircache")
}

func (aircacheOp) View(id string, t Text) Screen {
	switch id {
	case "aircache.home":
		s := Screen{Kind: "menu"}
		if p := loadPending(); p != nil {
			s.Items = append(s.Items, cmdItem("pending", "", p.Code+"\n"+p.Barcode))
		}
		s.Items = append(s.Items,
			navItem("aircache.code", t("ui.enter_code")),
			navItem("aircache.storno", t("ui.storno")),
			cmdItem("copy", "", t("ui.copy")),
			reportItem("report.aircache", "rsAircache", "detail", t("ui.summary")),
			reportItem("report.aircache", "rdAircache", "detail", t("ui.detail")),
		)
		return s
	case "aircache.code":
		return formScreen(Field{Name: "code", Label: t("ui.code"), Type: "text"})
	case "aircache.storno":
		return formScreen(Field{Name: "stid", Label: t("ui.txid"), Type: "text"})
	default:
		return infoScreen(id)
	}
}

func (aircacheOp) Submit(id string, v map[string]string, rt Runtime) {
	switch id {
	case "aircache.code":
		aircashCode(rt, v["code"])
	case "aircache.storno":
		aircashStorno(rt, v["stid"])
	}
}

func (aircacheOp) Tap(_ string, act Action, rt Runtime) bool {
	if act.Kind != ActCmd {
		return false
	}
	switch act.Target {
	case "pending":
		if p := loadPending(); p != nil {
			checkBarcode(rt, p.Barcode)
		}
		return true
	case "copy":
		aircashLast(rt)
		return true
	default:
		return false
	}
}

func acTmpl() *print.Template {
	b := print.NewBuilder()
	return b.Line().Text("AIRCASH").
		Line().Param("dts").
		Line().Param("mn").Text(",").Param("pa").
		Line().Text("Kod: ").Param("code").
		Line().Text("Barkod: ").Param("bc").
		Line().Param("amt").Text(" KM").
		Line().Text("ID transakcije").Line().Param("tid").
		Line().Template()
}

func loadPending() *pendingTx {
	b, err := os.ReadFile(aircacheFile)
	if err != nil {
		return nil
	}
	var p pendingTx
	if json.Unmarshal(b, &p) != nil || p.Code == "" {
		return nil
	}
	return &p
}

func savePending(code, bc string) {
	b, _ := json.Marshal(pendingTx{Code: code, Barcode: bc})
	_ = os.WriteFile(aircacheFile, b, 0o644)
}

func deletePending() { _ = os.Remove(aircacheFile) }

func handleAircash(rt Runtime, m map[string]string) {
	if m["err"] == "30" || m["err"] == "60" {
		rt.Toast(rt.T("toast.ok"))
		rt.Print(m, acTmpl(), "aircache", true)
		return
	}
	rt.Toast(rt.T(m["err"]))
}

func checkBarcode(rt Runtime, bc string) {
	pass := rt.Pass()
	h := rt.Socket().Put("com", "checkBarcodeStatusAc").Put("bc", bc).Put("pass", pass).Put("rc", strconv.Itoa(rt.NextRC()))
	rt.RunNet(h, func(m map[string]string) {
		if m["err"] == "42" {
			checkBarcode(rt, bc)
			return
		}
		deletePending()
		handleAircash(rt, m)
		rt.Refresh()
	}, func(err error) { checkBarcode(rt, bc) })
}

func aircashLast(rt Runtime) {
	pass := rt.Pass()
	h := rt.Socket().Put("com", "checkTransactionStatusAc").Put("pass", pass).Put("rc", strconv.Itoa(rt.NextRC()))
	rt.RunNet(h, func(m map[string]string) { handleAircash(rt, m) }, nil)
}

func aircashCode(rt Runtime, code string) {
	pass := rt.Pass()
	h := rt.Socket().Put("com", "checkCodeAc").Put("code", code).Put("pass", pass).Put("rc", strconv.Itoa(rt.NextRC()))
	rt.RunNet(h, func(m map[string]string) {
		if m["err"] != "" && m["err"] != "30" && m["err"] != "60" {
			rt.Toast(rt.T(m["err"]))
			return
		}
		bc := m["bc"]
		if bc == "" {
			bc = m["barcode"]
		}
		rt.Ask(rt.T("dialog.aircache"), func() {
			savePending(code, bc)
			ch := rt.Socket().Put("com", "confirmTransactionAc").Put("bc", bc).Put("pass", pass).Put("rc", strconv.Itoa(rt.NextRC()))
			rt.RunNet(ch, func(m2 map[string]string) {
				if m2["err"] == "42" {
					rt.Home()
					return
				}
				deletePending()
				if m2["err"] == "30" || m2["err"] == "60" {
					rt.Toast(rt.T("toast.ok"))
					rt.Print(m2, acTmpl(), "aircache", true)
					rt.Home()
					return
				}
				rt.Toast(rt.T(m2["err"]))
			}, nil)
		})
	}, nil)
}

func aircashStorno(rt Runtime, tid string) {
	pass := rt.Pass()
	h := rt.Socket().Put("com", "cancelTransactionAc").Put("stid", tid).Put("pass", pass).Put("rc", strconv.Itoa(rt.NextRC()))
	rt.RunNet(h, func(m map[string]string) {
		if m["err"] == "50" || m["err"] == "30" {
			rt.Toast(rt.T("toast.storno_ok"))
			rt.Print(m, acTmpl(), "aircache", true)
			rt.Home()
			return
		}
		rt.Toast(rt.T(m["err"]))
	}, nil)
}
