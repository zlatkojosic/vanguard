package fulgor

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// DecodeBody turns a Fulgor JSON body into the flat k=v map receipts still use.
func DecodeBody(raw []byte) (map[string]string, error) {
	if len(bytesTrim(raw)) == 0 {
		return map[string]string{}, nil
	}
	var top map[string]any
	if err := json.Unmarshal(raw, &top); err != nil {
		return nil, err
	}
	if _, hasCode := top["code"]; hasCode {
		out := map[string]string{}
		if n, ok := asInt(top["code"]); ok {
			out["err"] = strconv.Itoa(n)
		}
		switch data := top["data"].(type) {
		case map[string]any:
			flatten(data, out)
		case []any:
			flattenReportItems(data, out)
		}
		if out["err"] == "" {
			if n, ok := asInt(top["err"]); ok {
				out["err"] = strconv.Itoa(n)
			}
		}
		return out, nil
	}
	out := map[string]string{}
	flatten(top, out)
	return out, nil
}

func flatten(in map[string]any, out map[string]string) {
	if items, ok := in["data"].([]any); ok {
		flattenReportItems(items, out)
	}
	if products, ok := in["products"].([]any); ok {
		flattenStock(products, out)
	}
	if amounts, ok := in["voucherAmounts"].([]any); ok {
		for i, a := range amounts {
			out[fmt.Sprintf("va%d", i+1)] = asString(a)
		}
	}
	flattenFooter(in["footer"], out)
	flattenLampicaTotals(in, out)
	for k, v := range in {
		switch k {
		case "data", "footer", "products", "voucherAmounts", "operatorConfig":
			continue
		}
		key := jsonToSocket(k)
		if s := asString(v); s != "" || v == nil {
			if existing, ok := out[key]; !ok || existing == "" {
				out[key] = s
			}
		}
	}
	if out["servisi"] == "" && out["services"] != "" {
		out["servisi"] = out["services"]
	}
	if out["nSn"] == "" && out["newSerialNumber"] != "" {
		out["nSn"] = out["newSerialNumber"]
	}
	if out["rf"] == "" && out["dts"] != "" && (out["i1.1"] != "" || in["footer"] != nil) {
		out["rf"] = out["dts"]
	}
}

func flattenFooter(v any, out map[string]string) {
	footer, ok := v.(map[string]any)
	if !ok {
		return
	}
	if n, ok := asInt(footer["count"]); ok {
		out["sum1"] = strconv.Itoa(n)
	}
	if s := asString(footer["amount"]); s != "" {
		out["sum3"] = s
		out["sum2"] = s
	}
	if out["sum2"] == "" && out["sum1"] != "" {
		out["sum2"] = out["sum1"]
	}
}

func flattenStock(products []any, out map[string]string) {
	for i, item := range products {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		n := i + 1
		out[fmt.Sprintf("i%d.1", n)] = asString(row["productName"])
		out[fmt.Sprintf("i%d.2", n)] = asString(row["quantity"])
	}
}

func flattenLampicaTotals(in map[string]any, out map[string]string) {
	if _, has := in["data"].([]any); has {
		return
	}
	_, hasP := in["purchaseAmount"]
	_, hasV := in["voucherAmount"]
	if !hasP && !hasV {
		return
	}
	out["i1.1"] = asString(in["purchaseAmount"])
	out["i1.2"] = asString(in["voucherAmount"])
}

func flattenReportItems(items []any, out map[string]string) {
	for i, item := range items {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		n := i + 1
		a, b, c := reportCols(row)
		out[fmt.Sprintf("i%d.1", n)] = a
		out[fmt.Sprintf("i%d.2", n)] = b
		out[fmt.Sprintf("i%d.3", n)] = c
	}
}

func reportCols(row map[string]any) (string, string, string) {
	if _, ok := row["operaterName"]; ok {
		return asString(row["operaterName"]), asString(row["count"]), asString(row["amount"])
	}
	return asString(row["transactionId"]), asString(row["amount"]), asString(row["time"])
}

func jsonToSocket(k string) string {
	switch k {
	case "transactionId":
		return "tid"
	case "voucherCode":
		return "vc"
	case "dateTime":
		return "dts"
	case "merchantName":
		return "mn"
	case "posAddress":
		return "pa"
	case "phoneNumber":
		return "phN"
	case "serialCardNumber", "cardNumber":
		return "cn"
	case "newSerialNumber":
		return "nSn"
	case "services":
		return "servisi"
	case "amount":
		return "amt"
	case "barcode":
		return "bc"
	case "productCode":
		return "code"
	case "quantity":
		return "qu"
	case "stornTransactionId":
		return "stid"
	case "userTotalMoneyAmount":
		return "amt"
	case "userTotalPointsAmount":
		return "pts"
	case "forDate", "forDay":
		return "rh"
	case "expiryDateTime":
		return "exp"
	case "voucherSerialNumber":
		return "vsN"
	case "serialNumber":
		return "sn"
	case "productName":
		return "pn"
	case "text1":
		return "pm1"
	case "text2":
		return "pm2"
	case "text3":
		return "pm3"
	case "monthYear":
		return "mtr"
	case "requestCounter":
		return "rc"
	default:
		return k
	}
}

func asString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case json.Number:
		return t.String()
	case bool:
		return strconv.FormatBool(t)
	case int:
		return strconv.Itoa(t)
	default:
		return fmt.Sprint(t)
	}
}

func asInt(v any) (int, bool) {
	switch t := v.(type) {
	case float64:
		return int(t), true
	case int:
		return t, true
	case json.Number:
		n, err := t.Int64()
		return int(n), err == nil
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(t))
		return n, err == nil
	default:
		return 0, false
	}
}

func bytesTrim(b []byte) []byte {
	return []byte(strings.TrimSpace(string(b)))
}
