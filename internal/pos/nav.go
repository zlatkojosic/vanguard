package pos

import (
	"strconv"
	"strings"

	"nts/vanguard/internal/operators"
	"nts/vanguard/internal/print"
)

func (s *Session) Tap(id string) {
	if id == "" {
		return
	}
	act := operators.Decode(id)
	switch act.Kind {
	case operators.ActNav:
		s.open(act.Target)
	case operators.ActTheme:
		s.SetTheme(act.Target)
	case operators.ActLang:
		s.SetLang(act.Target)
	case operators.ActReport:
		title, cmd, kind := act.Target, act.Arg, act.Extra
		if title == "" {
			title = "title.report"
		}
		if kind == "" {
			kind = "detail"
		}
		s.mu.Lock()
		s.push(screen{ID: "report", Title: title, Args: map[string]string{"command": cmd, "kind": kind}})
		s.mu.Unlock()
		s.refresh()
	case operators.ActCmd:
		switch act.Target {
		case "reportShow":
			s.reportShow()
		case "reportPrint":
			s.reportPrint()
		case "upgrade":
			s.beginUpgrade()
		default:
			s.operatorTap(act)
		}
	default:
		s.operatorTap(act)
	}
}

func (s *Session) operatorTap(act operators.Action) {
	s.mu.Lock()
	cur := s.cur().ID
	s.mu.Unlock()
	operators.Tap(cur, act, s)
}

func (s *Session) open(id string) {
	s.mu.Lock()
	s.openLocked(id)
	s.mu.Unlock()
	s.refresh()
}

func (s *Session) openLocked(id string) {
	s.push(screen{ID: id, Title: s.titleFor(id), Args: map[string]string{}})
}

func (s *Session) titleFor(id string) string {
	switch id {
	case "login":
		return s.T("title.login")
	case "home":
		return s.T("title.home")
	case "theme":
		return s.T("title.theme")
	case "language":
		return s.T("title.language")
	case "system":
		return s.T("title.system")
	case "printer":
		return s.T("title.printer")
	case "margin":
		return s.T("title.margin")
	case "fiscal":
		return s.T("title.fiscal")
	case "keyboard":
		return s.T("title.keyboard")
	case "other":
		return s.T("title.other")
	case "upgrade":
		return s.T("title.upgrade")
	case "deregister":
		return s.T("title.deregister")
	case "report":
		return s.T("title.report")
	default:
		return operators.Title(id, s.T)
	}
}

func (s *Session) Submit(values map[string]string) {
	if values == nil {
		s.mu.Lock()
		values = cloneMap(s.form)
		s.mu.Unlock()
	} else {
		s.mu.Lock()
		s.form = cloneMap(values)
		s.mu.Unlock()
	}
	s.mu.Lock()
	id := s.cur().ID
	s.mu.Unlock()
	switch id {
	case "login":
		s.Login(values["pass"])
	case "system":
		s.saveSystem(values)
	case "printer":
		s.savePrinter(values)
	case "margin":
		s.saveMargin(values)
	case "fiscal":
		s.saveFiscal(values)
	case "keyboard":
		s.saveKeyboard(values)
	case "deregister":
		s.doDeregister(values["admin"])
	case "report":
		s.reportShow()
	default:
		operators.Submit(id, values, s)
	}
}

func cloneMap(m map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range m {
		out[k] = v
	}
	return out
}

func (s *Session) saveSystem(v map[string]string) {
	st := s.Pref
	st.PutString("userSerialNumber", v["userSerialNumber"])
	st.PutString("serverIp", v["serverIp"])
	p, _ := strconv.Atoi(v["serverPort"])
	st.PutInt("serverPort", p)
	transport := strings.ToLower(strings.TrimSpace(v["transport"]))
	if transport != "rest" {
		transport = "socket"
	}
	st.PutString("transport", transport)
	scheme := strings.ToLower(strings.TrimSpace(v["scheme"]))
	if scheme != "https" {
		scheme = "http"
	}
	st.PutString("scheme", scheme)
	st.PutString("storeIp", v["storeIp"])
	fp, _ := strconv.Atoi(v["storePort"])
	st.PutInt("storePort", fp)
	st.PutBool("userShortCut", v["userShortCut"] == "true")
	_ = st.Save()
	s.ResetREST()
	s.showToast(s.T("toast.saved"))
}

func (s *Session) savePrinter(v map[string]string) {
	if strings.TrimSpace(v["printerName"]) == "" {
		s.showToast(s.T("toast.printer_missing"))
		return
	}
	s.Pref.PutString("printerName", v["printerName"])
	_ = s.Pref.Save()
	s.showToast(s.T("toast.printer_picked", v["printerName"]))
	s.Back()
}

func (s *Session) saveMargin(v map[string]string) {
	st := s.Pref
	for _, k := range []string{"topMargin", "leftMargin", "rightMargin", "bottomMargin", "pageWidth"} {
		n, _ := strconv.Atoi(v[k])
		st.PutInt(k, n)
	}
	f, _ := strconv.ParseFloat(v["pageScale"], 64)
	st.PutFloat("pageScale", f)
	_ = st.Save()
	s.showToast(s.T("toast.saved_ok"))
}

func (s *Session) saveFiscal(v map[string]string) {
	st := s.Pref
	useF := v["useFiscal"] == "true"
	useS := v["useSerialPort"] == "true"
	if useF {
		useS = false
	}
	st.PutString("useFiscal", boolStr(useF))
	st.PutString("useSerialPort", boolStr(useS))
	st.PutString("fiscalPrinterType", v["fiscalPrinterType"])
	st.PutString("printerFolder", v["printerFolder"])
	n, _ := strconv.Atoi(v["printerRowLen"])
	st.PutInt("printerRowLen", n)
	st.PutString("fileExt", v["fileExt"])
	st.PutString("serialPrinterType", v["serialPrinterType"])
	st.PutString("comPort", v["comPort"])
	_ = st.Save()
	s.showToast(s.T("toast.saved_ok"))
}

func (s *Session) saveKeyboard(v map[string]string) {
	s.Pref.PutBool("useKeyboard", v["useKeyboard"] == "true")
	_ = s.Pref.Save()
	s.showToast(s.T("toast.saved_ok"))
}

func (s *Session) doDeregister(pass string) {
	if strings.TrimSpace(pass) != "admin@3" {
		s.showToast(s.T("toast.bad_password"))
		return
	}
	s.Pref.PutString("terminalSerialNumber", "")
	s.Pref.PutBool("registered", false)
	_ = s.Pref.Save()
	s.showToast(s.T("toast.deregistered"))
	s.Back()
}

func (s *Session) reportShow() {
	s.mu.Lock()
	sc := s.cur()
	fd := valueOr(s.form["fd"], todayDDMMYY())
	from := valueOr(s.form["fromTime"], "0000")
	to := valueOr(s.form["toTime"], "2359")
	cmd := sc.Args["command"]
	s.mu.Unlock()
	h := s.socket().Put("com", cmd).Put("fd", fd).Put("mtr", monthFrom(fd))
	if from != "" {
		h.Put("fromTime", from).Put("toTime", to)
	}
	s.runNet(h, func(m map[string]string) {
		m["fromTime"] = formatTime(from)
		m["toTime"] = formatTime(to)
		rows := [][]string{}
		for i := 1; ; i++ {
			a := m["i"+strconv.Itoa(i)+".1"]
			b := m["i"+strconv.Itoa(i)+".2"]
			c := m["i"+strconv.Itoa(i)+".3"]
			if a == "" || b == "" {
				break
			}
			rows = append(rows, []string{a, b, c})
		}
		s.mu.Lock()
		s.reportLast = m
		s.reportRows = rows
		s.reportSum = m["sum2"]
		s.reportAmt = m["sum3"]
		s.mu.Unlock()
		s.refresh()
	}, nil)
}

func (s *Session) reportPrint() {
	s.mu.Lock()
	last := s.reportLast
	kind := s.cur().Args["kind"]
	s.mu.Unlock()
	if last == nil {
		s.showToast(s.T("toast.show_first"))
		return
	}
	var tmpl *print.Template
	switch kind {
	case "week":
		tmpl = print.WeekReport()
	case "detail", "lampDay", "lampMonth":
		tmpl = print.DetailReport(s.T("report.detail"))
	default:
		tmpl = print.SumReport()
	}
	s.doPrint(last, tmpl, "true", true)
}

func monthFrom(ddMMyy string) string {
	if len(ddMMyy) >= 6 {
		return ddMMyy[2:]
	}
	return ddMMyy
}

func formatTime(hhmm string) string {
	if len(hhmm) == 4 {
		return hhmm[:2] + ":" + hhmm[2:]
	}
	return hhmm
}
