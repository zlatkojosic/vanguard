package fulgor

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type restReq struct {
	method string
	path   string
	body   map[string]any
}

func resolve(f map[string]string) (restReq, error) {
	com := f["com"]
	rc := atoi(f["rc"])
	switch com {
	case "rE":
		return post("/v2/eronet/voucher", voucherBody(f["amt"], rc, true)), nil
	case "crE":
		return get(fmt.Sprintf("/v2/eronet/voucher/%d", rc)), nil
	case "sEr":
		return post("/v2/eronet/voucher/last", emptyBody()), nil
	case "rdEr":
		return post("/v2/eronet/voucher/report/detail", reportBody(f)), nil

	case "rEt":
		return post("/v2/eronet/topup", topupBody(f, rc)), nil
	case "crEt":
		return get(fmt.Sprintf("/v2/eronet/topup/%d", rc)), nil
	case "sEt":
		return post("/v2/eronet/topup/last", emptyBody()), nil
	case "cEt":
		return post("/v2/eronet/topup/storno", stornoBody(f, rc)), nil
	case "rdEt":
		return post("/v2/eronet/topup/report/detail", reportBody(f)), nil

	case "rM":
		return post("/v2/mtel/topup", topupBody(f, rc)), nil
	case "crM":
		return get(fmt.Sprintf("/v2/mtel/topup/%d", rc)), nil
	case "sMt":
		return post("/v2/mtel/topup/last", emptyBody()), nil
	case "cMt":
		return post("/v2/mtel/topup/storno", stornoBody(f, rc)), nil
	case "rdMt":
		return post("/v2/mtel/topup/report/detail", reportBody(f)), nil

	case "rB":
		return post("/v2/bht/topup", topupBody(f, rc)), nil
	case "crB":
		return get(fmt.Sprintf("/v2/bht/topup/%d", rc)), nil
	case "sBh":
		return post("/v2/bht/topup/last", emptyBody()), nil
	case "cBh":
		return post("/v2/bht/topup/storno", stornoBody(f, rc)), nil
	case "rdBh":
		return post("/v2/bht/topup/report/detail", reportBody(f)), nil

	case "rBv":
		return post("/v2/bht/voucher", voucherBody(f["amt"], rc, true)), nil
	case "crBv":
		return get(fmt.Sprintf("/v2/bht/voucher/%d", rc)), nil
	case "bvs":
		return post("/v2/bht/voucher/last", emptyBody()), nil
	case "rdBv":
		return post("/v2/bht/voucher/report/detail", reportBody(f)), nil

	case "rHt":
		return post("/v2/haloo/topup", topupBody(f, rc)), nil
	case "crHt":
		return get(fmt.Sprintf("/v2/haloo/topup/%d", rc)), nil
	case "sHt":
		return post("/v2/haloo/topup/last", emptyBody()), nil
	case "cHt":
		return post("/v2/haloo/topup/storno", stornoBody(f, rc)), nil
	case "rdHt":
		return post("/v2/haloo/topup/report/detail", reportBody(f)), nil

	case "rHv":
		return post("/v2/haloo/voucher", voucherBody(f["amt"], rc, true)), nil
	case "crHv":
		return get(fmt.Sprintf("/v2/haloo/voucher/%d", rc)), nil
	case "sHv":
		return post("/v2/haloo/voucher/last", emptyBody()), nil
	case "rdHv":
		return post("/v2/haloo/voucher/report/detail", reportBody(f)), nil

	case "rNvt":
		return post("/v2/novotel/topup", topupBody(f, rc)), nil
	case "crNvt":
		return get(fmt.Sprintf("/v2/novotel/topup/%d", rc)), nil
	case "lNvt":
		return post("/v2/novotel/topup/last", emptyBody()), nil
	case "cNvt":
		return post("/v2/novotel/topup/storno", stornoBody(f, rc)), nil
	case "rdNvt":
		return post("/v2/novotel/topup/report/detail", reportBody(f)), nil

	case "rI":
		return post("/v2/izi/topup", topupBody(f, rc)), nil
	case "crI":
		return get(fmt.Sprintf("/v2/izi/topup/%d", rc)), nil
	case "sI":
		return post("/v2/izi/topup/last", emptyBody()), nil
	case "cI":
		return post("/v2/izi/topup/storno", stornoBody(f, rc)), nil
	case "rdI":
		return post("/v2/izi/topup/report/detail", reportBody(f)), nil

	case "rY":
		return post("/v2/happy/topup", topupBody(f, rc)), nil
	case "crY":
		return get(fmt.Sprintf("/v2/happy/topup/%d", rc)), nil
	case "sYh":
		return post("/v2/happy/topup/last", emptyBody()), nil
	case "cYh":
		return post("/v2/happy/topup/storno", stornoBody(f, rc)), nil
	case "rdYh":
		return post("/v2/happy/topup/report/detail", reportBody(f)), nil

	case "rYv":
		return post("/v2/happy/voucher", voucherBody(f["amt"], rc, true)), nil
	case "crYv":
		return get(fmt.Sprintf("/v2/happy/voucher/%d", rc)), nil
	case "Yvs":
		return post("/v2/happy/voucher/last", emptyBody()), nil
	case "rdvs":
		return post("/v2/happy/voucher/report/detail", reportBody(f)), nil

	case "sNv":
		return post("/v2/xpay/voucher", voucherBody(f["amt"], rc, false)), nil
	case "crNv":
		return get(fmt.Sprintf("/v2/xpay/voucher/%d", rc)), nil
	case "lNv":
		return post("/v2/xpay/voucher/last", emptyBody()), nil
	case "cXb":
		return post("/v2/xpay/voucher/storno", stornoBody(f, rc)), nil
	case "rdNv":
		return post("/v2/xpay/voucher/report/detail", reportBody(f)), nil

	case "startRechargeABon":
		return post("/v2/abon/voucher", voucherBody(f["amt"], rc, false)), nil
	case "startCheckABon":
		return get(fmt.Sprintf("/v2/abon/voucher/%d", rc)), nil
	case "lastABon":
		return post("/v2/abon/voucher/last", emptyBody()), nil
	case "detailReportABon":
		return post("/v2/abon/voucher/report/detail", reportBody(f)), nil

	case "rechargeSpiner":
		return post("/v2/kliker/voucher", voucherBody(f["amt"], rc, false)), nil
	case "checkSpiner":
		return get(fmt.Sprintf("/v2/kliker/voucher/%d", rc)), nil
	case "lastSpiner":
		return post("/v2/kliker/voucher/last", emptyBody()), nil
	case "chargeSpiner":
		return post("/v2/kliker/voucher/storno", stornoBody(f, rc)), nil
	case "rdSpiner":
		return post("/v2/kliker/voucher/report/detail", reportBody(f)), nil

	case "rAbc":
		return post("/v2/abc/voucher", voucherBody(f["amt"], rc, false)), nil
	case "scAbc":
		return get(fmt.Sprintf("/v2/abc/voucher/%d", rc)), nil
	case "lAbc":
		return post("/v2/abc/voucher/last", emptyBody()), nil
	case "cAbc":
		return post("/v2/abc/voucher/storno", stornoBody(f, rc)), nil
	case "rdAbc":
		return post("/v2/abc/voucher/report/detail", reportBody(f)), nil

	case "newVMozz":
		return post("/v2/mozz/voucher", voucherBody(f["amt"], rc, false)), nil
	case "checkVMozz":
		return get(fmt.Sprintf("/v2/mozz/voucher/%d", rc)), nil
	case "lastVMozz":
		return post("/v2/mozz/voucher/last", emptyBody()), nil
	case "chargeVMozz":
		return post("/v2/mozz/voucher/storno", stornoBody(f, rc)), nil
	case "rdMozz":
		return post("/v2/mozz/voucher/report/detail", reportBody(f)), nil

	case "rBnBn":
		return post("/api/v2/binbin/sell", map[string]any{
			"productCode": f["code"], "requestId": rc,
		}), nil
	case "crBnBn":
		return get(fmt.Sprintf("/api/v2/binbin/sell/%d", rc)), nil
	case "lBnBn":
		return post("/api/v2/binbin/last", emptyBody()), nil
	case "rdBnBn":
		return post("/api/v2/binbin/report/detail", reportBody(f)), nil

	case "psellSell":
		qty := atoi(f["qu"])
		return post("/api/v2/product/sell", map[string]any{
			"productCode": f["code"], "quantity": qty, "requestId": rc,
		}), nil
	case "psellCheck":
		return get(fmt.Sprintf("/api/v2/product/sales/%d", rc)), nil
	case "psellLast":
		return post("/api/v2/product/last", emptyBody()), nil
	case "psellStorno":
		return post("/api/v2/product/storno", stornoBody(f, rc)), nil
	case "rdPSell", "rdPSellStock":
		pos := first(f["tsN"], f["posIdent"])
		return get("/api/v2/product/store/" + url.PathEscape(pos)), nil

	case "rL":
		return post("/v2/lampica/purchase", map[string]any{
			"serialCardNumber": f["cn"], "amount": atof(f["amt"]), "requestId": rc,
		}), nil
	case "crL":
		return get(fmt.Sprintf("/v2/lampica/purchase/%d", rc)), nil
	case "sL":
		return post("/v2/lampica/purchase/last", emptyBody()), nil
	case "cL":
		return post("/v2/lampica/purchase/storno", stornoBody(f, rc)), nil
	case "rLv":
		return post("/v2/lampica/voucher", map[string]any{
			"serialCardNumber": f["cn"], "amount": atoi(f["amt"]), "requestId": rc,
		}), nil
	case "crLv":
		return get(fmt.Sprintf("/v2/lampica/voucher/%d", rc)), nil
	case "sLv":
		return post("/v2/lampica/voucher/last", emptyBody()), nil
	case "bL":
		return get("/v2/lampica/voucher/amounts/" + url.PathEscape(f["cn"])), nil
	case "rdL":
		return post("/v2/lampica/report/day", reportBody(f)), nil
	case "mdL":
		body := emptyBody()
		if f["mtr"] != "" {
			body["monthYear"] = f["mtr"]
		}
		return post("/v2/lampica/report/month", body), nil

	case "checkCodeAc":
		return post("/api/v1/aircash/checkCode", map[string]any{
			"code": f["code"], "signature": sigOrStub(f),
		}), nil
	case "confirmTransactionAc":
		return post("/api/v1/aircash/confirm", map[string]any{
			"barcode": first(f["bc"], f["barcode"]), "signature": sigOrStub(f),
		}), nil
	case "cancelTransactionAc":
		return post("/api/v1/aircash/cancel", map[string]any{
			"stornTransactionId": atoi64(f["stid"]), "signature": sigOrStub(f),
		}), nil
	case "checkBarcodeStatusAc":
		return post("/api/v1/aircash/checkBarcodeStatus", map[string]any{
			"barcode": first(f["bc"], f["barcode"]), "signature": sigOrStub(f),
		}), nil

	case "rsA":
		return post("/v2/recharge/report/summary", reportBody(f)), nil
	case "LogIn":
		return restReq{}, fmt.Errorf("login is not dispatched")
	case "gV", "cV", "rA", "rsW", "checkTransactionStatusAc", "rsAircache", "rdAircache":
		return restReq{}, fmt.Errorf("unsupported command %s", com)
	default:
		return restReq{}, fmt.Errorf("unknown command %s", com)
	}
}

func post(path string, body map[string]any) restReq {
	return restReq{method: "POST", path: path, body: body}
}

func get(path string) restReq {
	return restReq{method: "GET", path: path}
}

func emptyBody() map[string]any {
	return map[string]any{}
}

func voucherBody(amt string, rc int, integer bool) map[string]any {
	body := map[string]any{"requestId": rc}
	if integer {
		body["amount"] = atoi(amt)
	} else {
		body["amount"] = atof(amt)
	}
	return body
}

func topupBody(f map[string]string, rc int) map[string]any {
	return map[string]any{
		"phoneNumber": f["phN"], "amount": atof(f["amt"]), "requestId": rc,
	}
}

func stornoBody(f map[string]string, rc int) map[string]any {
	return map[string]any{
		"transactionId": f["stid"], "requestId": rc,
	}
}

func reportBody(f map[string]string) map[string]any {
	body := emptyBody()
	if fd := strings.TrimSpace(f["fd"]); fd != "" {
		body["forDay"] = fd
	}
	return body
}

func sigOrStub(f map[string]string) string {
	if s := f["signature"]; s != "" {
		return s
	}
	return "x"
}

func first(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func atoi(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

func atoi64(s string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return n
}

func atof(s string) float64 {
	n, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return n
}
