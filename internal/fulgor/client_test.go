package fulgor

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDecodeVoucherEnvelope(t *testing.T) {
	raw := []byte(`{"request":1,"code":30,"msg":"","data":{"err":30,"transactionId":1001,"voucherCode":"ABC","merchantName":"STUB","posAddress":"addr","dateTime":"20260823120000","amount":10}}`)
	m, err := DecodeBody(raw)
	if err != nil {
		t.Fatal(err)
	}
	if m["err"] != "30" || m["tid"] != "1001" || m["vc"] != "ABC" || m["mn"] != "STUB" || m["pa"] != "addr" || m["dts"] != "20260823120000" || m["amt"] != "10" {
		t.Fatalf("map %#v", m)
	}
}

func TestDecodeLogin(t *testing.T) {
	raw := []byte(`{"err":60,"services":"rE,rM","printable":"rE","appVer":"1.0","sessionId":"abc","posName":"STUB-POS","operatorConfig":{"currency":"KM","operators":{"xBon":{"amounts":[5,10]}}}}`)
	m, err := DecodeBody(raw)
	if err != nil {
		t.Fatal(err)
	}
	if m["err"] != "60" || m["servisi"] != "rE,rM" || m["sessionId"] != "abc" {
		t.Fatalf("map %#v", m)
	}
	if _, ok := m["operatorConfig"]; ok {
		t.Fatalf("nested operatorConfig should not flatten: %#v", m)
	}
	ov := ParseOperatorConfig(raw)
	if ov.Currency != "KM" || len(ov.Amounts["xBon"]) != 2 || ov.Amounts["xBon"][0] != 5 {
		t.Fatalf("overlay %#v", ov)
	}
}

func TestDecodeRegister(t *testing.T) {
	raw := []byte(`{"err":116,"newSerialNumber":"POS1120000"}`)
	m, err := DecodeBody(raw)
	if err != nil {
		t.Fatal(err)
	}
	if m["err"] != "116" || m["nSn"] != "POS1120000" {
		t.Fatalf("map %#v", m)
	}
}

func TestDecodeSumReport(t *testing.T) {
	raw := []byte(`{"merchantName":"STUB","posAddress":"addr","forDate":"230826","dateTime":"now","data":[{"operaterName":"Eronet","count":2,"amount":20.5}],"footer":{"count":2,"amount":20.5}}`)
	m, err := DecodeBody(raw)
	if err != nil {
		t.Fatal(err)
	}
	if m["i1.1"] != "Eronet" || m["i1.2"] != "2" || m["i1.3"] != "20.5" || m["sum1"] != "2" || m["sum2"] != "20.5" || m["sum3"] != "20.5" {
		t.Fatalf("map %#v", m)
	}
}

func TestDecodeVoucherFields(t *testing.T) {
	raw := []byte(`{"request":1,"code":30,"data":{"err":30,"transactionId":9,"voucherCode":"ABC","voucherSerialNumber":"SN1","expiryDateTime":"20270101","text1":"p1"}}`)
	m, err := DecodeBody(raw)
	if err != nil {
		t.Fatal(err)
	}
	if m["vsN"] != "SN1" || m["exp"] != "20270101" || m["pm1"] != "p1" {
		t.Fatalf("map %#v", m)
	}
}

func TestDecodeStockAndLampica(t *testing.T) {
	stock := []byte(`{"err":30,"merchantName":"STUB","products":[{"productName":"P1","quantity":4}]}`)
	m, err := DecodeBody(stock)
	if err != nil {
		t.Fatal(err)
	}
	if m["i1.1"] != "P1" || m["i1.2"] != "4" || m["mn"] != "STUB" {
		t.Fatalf("stock %#v", m)
	}
	lamp := []byte(`{"request":0,"code":30,"data":{"merchantName":"STUB","forDay":"260826","purchaseAmount":10.5,"voucherAmount":3}}`)
	m, err = DecodeBody(lamp)
	if err != nil {
		t.Fatal(err)
	}
	if m["err"] != "30" || m["i1.1"] != "10.5" || m["i1.2"] != "3" || m["rh"] != "260826" {
		t.Fatalf("lampica %#v", m)
	}
}

func TestResolveHappyAndStock(t *testing.T) {
	req, err := resolve(map[string]string{"com": "rY", "amt": "10", "phN": "061", "rc": "2"})
	if err != nil || req.path != "/v2/happy/topup" {
		t.Fatalf("happy topup: %+v %v", req, err)
	}
	req, err = resolve(map[string]string{"com": "rYv", "amt": "5", "rc": "3"})
	if err != nil || req.path != "/v2/happy/voucher" || req.body["amount"] != 5 {
		t.Fatalf("happy voucher: %+v %v", req, err)
	}
	req, err = resolve(map[string]string{"com": "rdPSell", "tsN": "POS1"})
	if err != nil || req.method != "GET" || req.path != "/api/v2/product/store/POS1" {
		t.Fatalf("stock: %+v %v", req, err)
	}
}

func TestResolveVoucherAndStorno(t *testing.T) {
	req, err := resolve(map[string]string{"com": "rE", "tsN": "POS1", "amt": "10", "rc": "3"})
	if err != nil || req.method != "POST" || req.path != "/v2/eronet/voucher" {
		t.Fatalf("sell: %+v %v", req, err)
	}
	if req.body["amount"] != 10 || req.body["requestId"] != 3 {
		t.Fatalf("body %#v", req.body)
	}
	if _, ok := req.body["posIdent"]; ok {
		t.Fatalf("post-login body should not repeat posIdent: %#v", req.body)
	}
	req, err = resolve(map[string]string{"com": "cEt", "tsN": "POS1", "stid": "1001", "rc": "4"})
	if err != nil || req.path != "/v2/eronet/topup/storno" || req.body["transactionId"] != "1001" {
		t.Fatalf("storno: %+v %v", req, err)
	}
}

func TestClientLoginRegisterDispatch(t *testing.T) {
	var lastPath, lastUser, lastSession, lastAuth string
	var lastBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastPath = r.URL.Path
		lastUser = r.Header.Get("userId")
		lastSession = r.Header.Get("sessionId")
		lastAuth = r.Header.Get("Authorization")
		if r.Body != nil {
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &lastBody)
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v2/register":
			_, _ = w.Write([]byte(`{"err":116,"newSerialNumber":"POS1120000"}`))
		case "/v2/login":
			if lastUser != "" || r.Header.Get("token") != "" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"err":400}`))
				return
			}
			_, _ = w.Write([]byte(`{"err":60,"services":"rE,rM","printable":"rE","appVer":"1.0","sessionId":"sid-1","operatorConfig":{"currency":"KM","operators":{"ero":{"amounts":[2,5,10]}}}}`))
		case "/v2/eronet/voucher":
			if lastSession != "sid-1" || lastAuth != "Bearer sid-1" {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"request":0,"code":92,"msg":"not authenticated","data":null}`))
				return
			}
			_, _ = w.Write([]byte(`{"request":1,"code":30,"msg":"","data":{"err":30,"transactionId":9,"voucherCode":"V"}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL})
	reg, err := c.Register("POS1")
	if err != nil || reg["nSn"] != "POS1120000" || lastUser != "" {
		t.Fatalf("register %#v %v user=%s", reg, err, lastUser)
	}
	login, err := c.Login("POS1120000")
	if err != nil || login["err"] != "60" || c.Session() != "sid-1" || login["servisi"] != "rE,rM" {
		t.Fatalf("login %#v %v sid=%s", login, err, c.Session())
	}
	if ov := c.Operators(); ov.Currency != "KM" || ov.Amounts["ero"][0] != 2 {
		t.Fatalf("operators %#v", ov)
	}
	sold, err := c.Dispatch(map[string]string{"com": "rE", "tsN": "POS1120000", "amt": "10", "rc": "1"})
	if err != nil || sold["err"] != "30" || sold["tid"] != "9" || sold["vc"] != "V" {
		t.Fatalf("sell %#v %v path=%s", sold, err, lastPath)
	}
	if lastPath != "/v2/eronet/voucher" || lastSession != "sid-1" || lastAuth != "Bearer sid-1" {
		t.Fatalf("path=%s session=%s auth=%s", lastPath, lastSession, lastAuth)
	}
	reg, err = c.Register("POS1120000")
	if err != nil || lastUser != "" || lastSession != "sid-1" {
		t.Fatalf("register-with-session user=%s session=%s %#v %v", lastUser, lastSession, reg, err)
	}
}

func TestSkipTLSVerifyLoopbackHTTPS(t *testing.T) {
	if !skipTLSVerify("https://127.0.0.1:9494") || !skipTLSVerify("https://localhost:9494") {
		t.Fatal("expected skip for loopback https")
	}
	if skipTLSVerify("http://127.0.0.1:9494") || skipTLSVerify("https://pos.example:9494") {
		t.Fatal("did not expect skip for http or remote https")
	}
}
