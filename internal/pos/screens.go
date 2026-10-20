package pos

import (
	"os"
	"strconv"
	"strings"
	"time"

	"nts/vanguard/internal/i18n"
	"nts/vanguard/internal/operators"
	"nts/vanguard/internal/store"
)

func (s *Session) buildUI() UI {
	sc := s.cur()
	u := UI{
		Title:    s.titleFor(sc.ID),
		AppTitle: s.appTitle(),
		Theme:    s.theme,
		Accent:   AccentFor(s.theme),
		Message:  s.message,
		Toast:    s.toast,
		Busy:     s.busy,
		Confirm:  s.confirm,
		Keyboard: s.Pref.GetBool("useKeyboard", false),
		Columns:  4,
		Lang:     s.Tr.Lang(),
		Labels:   s.chromeLabels(),
	}
	if sc.ID == "report" && sc.Title != "" {
		u.Title = s.T(sc.Title)
	}
	switch sc.ID {
	case "login":
		s.uiLogin(&u)
	case "home":
		s.uiHome(&u)
	case "theme":
		s.uiTheme(&u)
	case "language":
		s.uiLanguage(&u)
	case "system":
		s.uiSystem(&u)
	case "printer":
		s.uiPrinter(&u)
	case "margin":
		s.uiMargin(&u)
	case "fiscal":
		s.uiFiscal(&u)
	case "keyboard":
		s.uiKeyboard(&u)
	case "other":
		s.uiOther(&u)
	case "report":
		s.uiReport(&u, sc)
	case "upgrade":
		s.uiUpgrade(&u)
	case "deregister":
		s.uiDeregister(&u)
	default:
		applyScreen(&u, operators.ApplyOverlay(sc.ID, operators.View(sc.ID, s.T), s.opCfg))
	}
	return u
}

func (s *Session) chromeLabels() map[string]string {
	keys := []string{"exit", "home", "back", "ok", "yes", "no", "confirm", "busy", "login", "apps", "settings", "show", "print"}
	m := make(map[string]string, len(keys))
	for _, k := range keys {
		m[k] = s.T("chrome." + k)
	}
	return m
}

func (s *Session) uiLogin(u *UI) {
	u.Kind = "tabs"
	u.Login = true
	if s.tab == "" {
		s.tab = "login"
	}
	u.Tab = s.tab
	u.Items = s.settingsItems()
	u.Fields = []Field{{Name: "pass", Label: s.T("ui.password"), Type: "password", Placeholder: s.T("ui.password_ph")}}
}

func (s *Session) uiHome(u *UI) {
	u.Kind = "tabs"
	if s.tab == "" || s.tab == "login" {
		s.tab = "apps"
	}
	u.Tab = s.tab
	filtered := filterMetas(s.apps, s.services)
	items := make([]Item, 0, len(filtered)+2)
	if s.pendingUpgrade() != "" {
		items = append(items, Item{ID: "go:upgrade", Title: s.T("ui.upgrade_now")})
	}
	for _, m := range filtered {
		items = append(items, Item{ID: "go:" + m.Activity, Title: m.Title})
	}
	items = append(items, Item{ID: "go:other", Title: s.T("ui.other")})
	u.Items = items
	if u.Tab == "settings" {
		u.Items = s.settingsItems()
	}
}

func (s *Session) uiUpgrade(u *UI) {
	if s.upgrading {
		u.Kind = "info"
		u.Info = s.T("ui.upgrade")
		return
	}
	u.Kind = "menu"
	u.Info = s.T("ui.upgrade_available")
	u.Items = []Item{{ID: "cmd:upgrade", Title: s.T("ui.upgrade_now")}}
}

func (s *Session) settingsItems() []Item {
	return []Item{
		{ID: "go:theme", Title: s.T("ui.appearance")},
		{ID: "go:language", Title: s.T("ui.language")},
		{ID: "go:system", Title: s.T("ui.application")},
		{ID: "go:printer", Title: s.T("ui.printer")},
		{ID: "go:margin", Title: s.T("ui.margins")},
		{ID: "go:fiscal", Title: s.T("ui.fiscal")},
		{ID: "go:keyboard", Title: s.T("ui.keyboard")},
		{ID: "go:deregister", Title: s.T("ui.deregister")},
	}
}

func applyScreen(u *UI, sc operators.Screen) {
	u.Kind = sc.Kind
	if sc.Columns > 0 {
		u.Columns = sc.Columns
	}
	u.Info = sc.Info
	u.Summary = sc.Summary
	if len(sc.Items) > 0 {
		u.Items = make([]Item, len(sc.Items))
		for i, it := range sc.Items {
			u.Items[i] = Item{ID: it.ID, Title: it.Title, Color: it.Color}
		}
	}
	if len(sc.Fields) > 0 {
		u.Fields = make([]Field, len(sc.Fields))
		for i, f := range sc.Fields {
			u.Fields[i] = Field{Name: f.Name, Label: f.Label, Type: f.Type, Value: f.Value, Options: f.Options, Placeholder: f.Placeholder, MaxLen: f.MaxLen, Numeric: f.Numeric}
		}
	}
	if sc.Table != nil {
		u.Table = &Table{Heads: sc.Table.Heads, Rows: sc.Table.Rows, Actions: sc.Table.Actions}
	}
}

func filterMetas(metas []operators.Meta, services string) []operators.Meta {
	parts := strings.Split(services, ",")
	has := map[string]bool{}
	for _, p := range parts {
		has[strings.TrimSpace(p)] = true
	}
	var out []operators.Meta
	for _, m := range metas {
		if m.Always {
			out = append(out, m)
			continue
		}
		for _, t := range m.Tags {
			if has[t] {
				out = append(out, m)
				break
			}
		}
	}
	return out
}

func (s *Session) uiTheme(u *UI) {
	u.Kind = "theme"
	u.Info = s.T("ui.pick_theme")
	for _, t := range Themes() {
		u.Items = append(u.Items, Item{ID: "theme:" + t.ID, Title: t.Label, Color: t.Accent})
	}
}

func (s *Session) uiLanguage(u *UI) {
	u.Kind = "theme"
	u.Info = s.T("ui.pick_language")
	for _, l := range i18n.Languages() {
		u.Items = append(u.Items, Item{ID: "lang:" + l.ID, Title: l.Label})
	}
}

func (s *Session) uiSystem(u *UI) {
	st := s.Pref
	u.Kind = "form"
	u.Fields = []Field{
		{Name: "userSerialNumber", Label: s.T("ui.terminal_sn"), Type: "text", Value: st.GetString("userSerialNumber", "")},
		{Name: "serverIp", Label: s.T("ui.server_ip"), Type: "text", Value: st.GetString("serverIp", "")},
		{Name: "serverPort", Label: s.T("ui.server_port"), Type: "text", Value: strconv.Itoa(st.GetInt("serverPort", 0))},
		{Name: "transport", Label: s.T("ui.transport"), Type: "select", Value: transportOf(st), Options: []string{"socket", "rest"}},
		{Name: "scheme", Label: s.T("ui.scheme"), Type: "select", Value: schemeOf(st), Options: []string{"http", "https"}},
		{Name: "storeIp", Label: s.T("ui.ftp_ip"), Type: "text", Value: st.GetString("storeIp", "")},
		{Name: "storePort", Label: s.T("ui.ftp_port"), Type: "text", Value: strconv.Itoa(st.GetInt("storePort", 0))},
		{Name: "userShortCut", Label: s.T("ui.shortcuts"), Type: "check", Value: boolStr(st.GetBool("userShortCut", false))},
	}
}

func transportOf(st store.Store) string {
	t := strings.ToLower(st.GetString("transport", "socket"))
	if t == "rest" || t == "http" || t == "fulgor" {
		return "rest"
	}
	return "socket"
}

func schemeOf(st store.Store) string {
	if st.GetString("scheme", "http") == "https" {
		return "https"
	}
	return "http"
}

func (s *Session) uiPrinter(u *UI) {
	u.Kind = "form"
	u.Fields = []Field{
		{Name: "printerName", Label: s.T("ui.pick_printer"), Type: "select", Value: s.Pref.GetString("printerName", ""), Options: s.Printers()},
	}
}

func (s *Session) uiMargin(u *UI) {
	st := s.Pref
	u.Kind = "form"
	u.Fields = []Field{
		{Name: "topMargin", Label: s.T("ui.top_margin"), Type: "text", Value: strconv.Itoa(st.GetInt("topMargin", 0))},
		{Name: "leftMargin", Label: s.T("ui.left_margin"), Type: "text", Value: strconv.Itoa(st.GetInt("leftMargin", 0))},
		{Name: "rightMargin", Label: s.T("ui.right_margin"), Type: "text", Value: strconv.Itoa(st.GetInt("rightMargin", 0))},
		{Name: "bottomMargin", Label: s.T("ui.bottom_margin"), Type: "text", Value: strconv.Itoa(st.GetInt("bottomMargin", 0))},
		{Name: "pageWidth", Label: s.T("ui.page_width"), Type: "text", Value: strconv.Itoa(st.GetInt("pageWidth", 0))},
		{Name: "pageScale", Label: s.T("ui.page_scale"), Type: "text", Value: strconv.FormatFloat(st.GetFloat("pageScale", 0.8), 'f', -1, 64)},
	}
}

func (s *Session) uiFiscal(u *UI) {
	st := s.Pref
	ports := hardwarePorts(s)
	u.Kind = "form"
	u.Fields = []Field{
		{Name: "useFiscal", Label: s.T("ui.use_fiscal"), Type: "check", Value: st.GetString("useFiscal", "false")},
		{Name: "fiscalPrinterType", Label: s.T("ui.printer_type"), Type: "select", Value: st.GetString("fiscalPrinterType", "Tring"), Options: []string{"TringT202", "Tring", "Tremol", "FP&TM", "MEP1000", "Elink"}},
		{Name: "printerFolder", Label: s.T("ui.printer_folder"), Type: "text", Value: st.GetString("printerFolder", "")},
		{Name: "printerRowLen", Label: s.T("ui.row_len"), Type: "text", Value: strconv.Itoa(st.GetInt("printerRowLen", 30))},
		{Name: "fileExt", Label: s.T("ui.file_ext"), Type: "text", Value: st.GetString("fileExt", "")},
		{Name: "useSerialPort", Label: s.T("ui.use_serial"), Type: "check", Value: st.GetString("useSerialPort", "false")},
		{Name: "serialPrinterType", Label: s.T("ui.serial_type"), Type: "select", Value: st.GetString("serialPrinterType", "Default"), Options: []string{"Default", "Secunda"}},
		{Name: "comPort", Label: s.T("ui.com_port"), Type: "select", Value: st.GetString("comPort", ""), Options: ports},
	}
}

func hardwarePorts(s *Session) []string {
	return s.SerialPorts()
}

func (s *Session) uiKeyboard(u *UI) {
	u.Kind = "form"
	u.Fields = []Field{
		{Name: "useKeyboard", Label: s.T("ui.use_keyboard"), Type: "check", Value: boolStr(s.Pref.GetBool("useKeyboard", false))},
	}
}

func (s *Session) uiOther(u *UI) {
	u.Kind = "menu"
	u.Items = []Item{
		{ID: operators.Encode(operators.Action{Kind: operators.ActReport, Target: "report.week", Arg: "rsW", Extra: "week"}), Title: s.T("ui.week")},
		{ID: operators.Encode(operators.Action{Kind: operators.ActReport, Target: "report.sum", Arg: "rsA", Extra: "sum"}), Title: s.T("ui.summary")},
	}
}

func (s *Session) uiReport(u *UI, sc screen) {
	u.Kind = "report"
	kind := sc.Args["kind"]
	heads := []string{s.T("ui.col_id"), s.T("ui.col_amount"), s.T("ui.col_time")}
	if kind == "sum" {
		heads = []string{s.T("ui.col_service"), s.T("ui.col_count"), s.T("ui.col_amount")}
	}
	u.Fields = []Field{
		{Name: "fd", Label: s.T("ui.date"), Type: "date", Value: dateToInput(valueOr(s.form["fd"], todayDDMMYY()))},
		{Name: "fromTime", Label: s.T("ui.from"), Type: "time", Value: formatTime(valueOr(s.form["fromTime"], "0000"))},
		{Name: "toTime", Label: s.T("ui.to"), Type: "time", Value: formatTime(valueOr(s.form["toTime"], "2359"))},
	}
	u.Table = &Table{Heads: heads, Rows: s.reportRows}
	u.Summary = s.T("ui.tx_count", s.reportSum)
	if s.reportAmt != "" {
		u.Summary += "\n" + s.T("ui.tx_sum", s.reportAmt)
	}
}

func (s *Session) uiDeregister(u *UI) {
	u.Kind = "form"
	u.Fields = []Field{{Name: "admin", Label: s.T("ui.admin_pass"), Type: "password"}}
}

func (s *Session) appTitle() string {
	return "N POS Z v" + s.Pref.GetString("appVer", Version)
}

func todayDDMMYY() string { return time.Now().Format("020106") }

// dateToInput converts ddMMyy to the yyyy-MM-dd value of an HTML date input.
func dateToInput(ddMMyy string) string {
	t, err := time.Parse("020106", ddMMyy)
	if err != nil {
		return ""
	}
	return t.Format("2006-01-02")
}

func boolStr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func valueOr(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func LoadProperties(st store.Store) {
	b, err := os.ReadFile("app.properties")
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		switch k {
		case "server":
			st.PutString("serverIp", v)
		case "port":
			n, _ := strconv.Atoi(v)
			st.PutInt("serverPort", n)
		case "useFiscal":
			st.PutString("useFiscal", v)
		case "useSerial":
			st.PutString("useSerialPort", v)
		case "printerName":
			st.PutString("printerName", v)
		case "prefix":
			st.PutString("prefix", v)
		case "pageWidth":
			n, _ := strconv.Atoi(v)
			st.PutInt("pageWidth", n)
		case "pageScale":
			f, _ := strconv.ParseFloat(v, 64)
			st.PutFloat("pageScale", f)
		case "transport":
			st.PutString("transport", v)
		case "scheme":
			st.PutString("scheme", v)
		}
	}
	_ = st.Save()
}
