package operators

import (
	"strconv"
	"strings"
)

const DefaultCurrency = "KM"

// Overlay is voucher amounts + currency from Fulgor. Empty fields keep catalog defaults.
type Overlay struct {
	Currency string
	Amounts  map[string][]int
}

func (o Overlay) CurrencyOr(def string) string {
	c := strings.TrimSpace(o.Currency)
	if c == "" {
		return def
	}
	return c
}

func (o Overlay) amountsFor(service string, fallback []int) []int {
	if o.Amounts != nil {
		if a, ok := o.Amounts[service]; ok && len(a) > 0 {
			return a
		}
	}
	return fallback
}

// ApplyOverlay replaces voucher amount tiles when Fulgor sent a catalog for this service.
func ApplyOverlay(id string, sc Screen, ov Overlay) Screen {
	f, _, ok := FlowByRoute(id)
	if !ok || f == nil || f.Amounts == nil {
		return sc
	}
	if !strings.HasSuffix(id, ".recharge") && !strings.HasSuffix(id, ".vrecharge") {
		return sc
	}
	amounts := ov.amountsFor(f.Service, f.Amounts)
	cur := ov.CurrencyOr(DefaultCurrency)
	out := Screen{Kind: "menu"}
	for _, a := range amounts {
		out.Items = append(out.Items, buyItem(strconv.Itoa(a), strconv.Itoa(a)+" "+cur))
	}
	return out
}
