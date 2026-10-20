package operators

import "testing"

type toastRT struct {
	Runtime
	toasts []string
	asked  bool
}

func (r *toastRT) Toast(m string)              { r.toasts = append(r.toasts, m) }
func (r *toastRT) T(k string, _ ...any) string { return k }
func (r *toastRT) Ask(_ string, _ func())      { r.asked = true }

func TestValidDigits(t *testing.T) {
	cases := []struct {
		in   string
		max  int
		want bool
	}{
		{"0651234567", 10, true},
		{"387651234678998777777888", 10, false},
		{"06512345678", 10, false},
		{"", 10, false},
		{"5.5", 0, false},
		{"5,5", 0, false},
		{"12", 0, true},
		{"12345678901234", 14, true},
		{"123456789012345", 14, false},
	}
	for _, c := range cases {
		if got := validDigits(c.in, c.max); got != c.want {
			t.Errorf("validDigits(%q,%d)=%v want %v", c.in, c.max, got, c.want)
		}
	}
}

func TestEronetTopupValidation(t *testing.T) {
	cases := []struct {
		name, phone, amt string
		wantToast        string
	}{
		{"too long phone", "387651234678998777777888", "5", "toast.invalid_phone"},
		{"decimal amount", "0651234567", "5.5", "toast.invalid_amount"},
		{"ok", "0651234567", "5", ""},
	}
	for _, c := range cases {
		rt := &toastRT{}
		runTopup("eronet.topup.recharge", map[string]string{"phN": c.phone, "amt": c.amt}, rt)
		if c.wantToast == "" {
			if !rt.asked || len(rt.toasts) != 0 {
				t.Errorf("%s: expected confirm, got toasts %v", c.name, rt.toasts)
			}
			continue
		}
		if rt.asked || len(rt.toasts) != 1 || rt.toasts[0] != c.wantToast {
			t.Errorf("%s: toasts=%v asked=%v", c.name, rt.toasts, rt.asked)
		}
	}
}

func TestEronetChargeValidation(t *testing.T) {
	rt := &toastRT{}
	runCharge("eronet.topup.charge", "123456789012345", rt)
	if len(rt.toasts) != 1 || rt.toasts[0] != "toast.invalid_stid" {
		t.Fatalf("toasts=%v", rt.toasts)
	}
}

func TestEronetFieldLimits(t *testing.T) {
	s := View("eronet.topup.recharge", Keys)
	if s.Fields[0].MaxLen != 10 || s.Fields[0].Numeric != "int" || s.Fields[1].Numeric != "int" {
		t.Fatalf("fields: %+v", s.Fields)
	}
	c := View("eronet.topup.charge", Keys)
	if c.Fields[0].MaxLen != 14 {
		t.Fatalf("storno field: %+v", c.Fields)
	}
}
