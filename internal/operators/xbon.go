package operators

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"nts/vanguard/internal/print"
)

type xbonTx struct {
	ID, Phone, Amount string
	Inserted          time.Time
}

const xbonFile = "xbonTransactions.json"

type xbonSMSOp struct{}

func (xbonSMSOp) Owns(id string) bool {
	return strings.HasPrefix(id, "xbon.sms")
}

func (xbonSMSOp) View(id string, t Text) Screen {
	if strings.HasSuffix(id, ".charge") {
		return xbonChargeTable(t)
	}
	return flowScreen(id, t)
}

func (xbonSMSOp) Submit(id string, form map[string]string, rt Runtime) {
	if strings.HasSuffix(id, ".recharge") {
		runTopup(id, form, rt)
	}
}

func (xbonSMSOp) Tap(id string, act Action, rt Runtime) bool {
	switch act.Kind {
	case ActLast:
		runLast(act.Target, rt)
		return true
	case ActBuy:
		runAmountBuy(id, act.Target, rt, nil)
		return true
	case ActCmd:
		if act.Target == "storno" {
			xbonStorno(rt, act.Arg)
			return true
		}
	}
	return false
}

func xbonChargeTable(t Text) Screen {
	txs := loadXBon()
	rows := make([][]string, 0, len(txs))
	actions := make([]string, 0, len(txs))
	for _, tx := range txs {
		rows = append(rows, []string{tx.Phone, tx.Amount, tx.ID, tx.Inserted.Format("02.01.2006 15:04"), t("ui.storn")})
		actions = append(actions, Encode(Action{Kind: ActCmd, Target: "storno", Arg: tx.ID}))
	}
	return Screen{
		Kind: "table",
		Table: &Table{
			Heads:   []string{t("ui.col_phone"), t("ui.col_amount"), t("ui.col_id"), t("ui.col_date"), t("ui.col_storno")},
			Rows:    rows,
			Actions: actions,
		},
	}
}

func loadXBon() []xbonTx {
	b, err := os.ReadFile(xbonFile)
	if err != nil {
		return nil
	}
	var all []xbonTx
	_ = json.Unmarshal(b, &all)
	cut := time.Now().Add(-10 * time.Minute)
	var live []xbonTx
	for _, t := range all {
		if t.Inserted.After(cut) {
			live = append(live, t)
		}
	}
	return live
}

func saveXBon(all []xbonTx) {
	b, _ := json.Marshal(all)
	_ = os.WriteFile(xbonFile, b, 0o644)
}

func xbonStorno(rt Runtime, tid string) {
	f, _, ok := FlowByRoute("xbon.sms.charge")
	if !ok || f == nil {
		return
	}
	h := rt.Socket().Put("com", f.Charge).Put("stid", tid).Put("rc", strconv.Itoa(rt.NextRC()))
	rt.RunNet(h, func(m map[string]string) {
		if m["err"] == "50" {
			rt.Toast(rt.T("toast.storno_ok"))
			rt.Print(m, print.Receipt("XBon", "charge"), "xBon", true)
			var rest []xbonTx
			for _, o := range loadXBon() {
				if o.ID != tid {
					rest = append(rest, o)
				}
			}
			saveXBon(rest)
			rt.Home()
			return
		}
		rt.Toast(rt.T(m["err"]))
	}, nil)
}
