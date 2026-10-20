package pos

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"nts/vanguard/internal/card"
	"nts/vanguard/internal/fulgor"
	"nts/vanguard/internal/hardware"
	"nts/vanguard/internal/i18n"
	"nts/vanguard/internal/netproto"
	"nts/vanguard/internal/operators"
	"nts/vanguard/internal/print"
	"nts/vanguard/internal/store"
	"nts/vanguard/internal/upgrade"
)

const Version = "1.0"

type screen struct {
	ID    string
	Title string
	Args  map[string]string
}

type Session struct {
	Pref    store.Store
	Mem     store.Store
	Tr      *i18n.Translator
	OnUI    func(UI)
	Quit    func()
	LockDir string

	mu         sync.Mutex
	stack      []screen
	tab        string
	theme      string
	message    string
	toast      string
	busy       bool
	confirm    string
	confirmYes func()
	form       map[string]string
	allowBuy   bool
	reportRows [][]string
	reportLast map[string]string
	reportSum  string
	reportAmt  string
	apps       []operators.Meta
	services   string
	opCfg      operators.Overlay
	cards      *card.Mux
	stopActive chan struct{}
	restClient *fulgor.Client
	upgrading  bool
	upgradeOffered bool
	// installUpgrade, if set, replaces FTP download (tests).
	installUpgrade func(st store.Store, ver, lockDir string) error
}

func New(pref store.Store) *Session {
	tr := i18n.New()
	tr.SetLang(pref.GetString("language", i18n.DefaultLang))
	s := &Session{
		Pref:  pref,
		Mem:   store.NewMemory(),
		Tr:    tr,
		theme: pref.GetString("theme-name", "theme3.css"),
		tab:   "login",
		apps:  operators.Metas(),
		cards: card.NewMux(card.NewEng(), card.NewLampica()),
		form:  map[string]string{},
	}
	s.stack = []screen{{ID: "login", Title: "Prijava"}}
	if pref.GetString("appVer", "") == "" {
		pref.PutString("appVer", Version)
		_ = pref.Save()
	}
	return s
}

func (s *Session) Start() {
	s.refresh()
	s.startActive()
	s.handleUpgrade()
}

func (s *Session) lockDir() string {
	if s.LockDir != "" {
		return s.LockDir
	}
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "vanguard")
}

func (s *Session) quit() {
	if s.Quit != nil {
		s.Quit()
		return
	}
	os.Exit(0)
}

func (s *Session) showUpgradeUI() {
	s.showUpgradeProgress()
}

func (s *Session) handleUpgrade() {
	if skipSelfUpdate() {
		log.Println("upgrade skip: wails dev does not replace the running process")
		return
	}
	p, err := upgrade.Current()
	if err != nil {
		log.Println("upgrade paths", err)
		return
	}
	st := upgrade.ReadStatus(p.Status)
	switch st.State {
	case upgrade.StateApplied:
		if st.Version != "" {
			s.commitVersion(st.Version)
		}
		return
	case upgrade.StateFailed:
		fail := st.Version
		if fail == "" {
			fail = s.Pref.GetString("newAppVer", "")
		}
		s.Pref.PutString("upgradeFailVer", fail)
		s.Pref.PutString("newAppVer", "")
		_ = s.Pref.Save()
		upgrade.ClearStatus(p.Status)
		s.mu.Lock()
		s.message = s.T("toast.upgrade_rollback")
		s.mu.Unlock()
		s.refresh()
		return
	case upgrade.StateReady:
		s.showUpgradeProgress()
		go func() {
			if err := upgrade.StartUpdater(p, st.Version, os.Getpid(), s.lockDir()); err != nil {
				s.showToast(s.T("toast.upgrade_fail", err.Error()))
				s.Home()
				return
			}
			time.Sleep(400 * time.Millisecond)
			s.quit()
		}()
		return
	}
}

func (s *Session) pendingUpgrade() string {
	nv := s.Pref.GetString("newAppVer", "")
	if nv == "" || nv == s.Pref.GetString("appVer", Version) {
		return ""
	}
	if nv == s.Pref.GetString("upgradeFailVer", "") {
		return ""
	}
	return nv
}

func (s *Session) showUpgradePrompt() {
	s.mu.Lock()
	s.push(screen{ID: "upgrade", Title: s.T("title.upgrade")})
	s.mu.Unlock()
	s.refresh()
}

func (s *Session) showUpgradeProgress() {
	s.mu.Lock()
	s.stack = []screen{{ID: "upgrade", Title: s.T("title.upgrade")}}
	s.mu.Unlock()
	s.refresh()
}

func (s *Session) offerUpgrade() {
	if skipSelfUpdate() {
		return
	}
	if s.pendingUpgrade() == "" {
		return
	}
	s.mu.Lock()
	if s.upgrading || s.upgradeOffered || s.cur().ID == "upgrade" || s.cur().ID == "login" {
		s.mu.Unlock()
		return
	}
	if s.cur().ID != "home" {
		s.mu.Unlock()
		return
	}
	s.upgradeOffered = true
	s.mu.Unlock()
	s.showUpgradePrompt()
}

func (s *Session) beginUpgrade() {
	if skipSelfUpdate() {
		s.showToast(s.T("toast.upgrade_dev"))
		return
	}
	if s.pendingUpgrade() == "" {
		return
	}
	if !s.ftpConfigured() {
		s.showToast(s.T("toast.upgrade_ftp"))
		return
	}
	s.startPendingDownload()
}

func (s *Session) ftpConfigured() bool {
	return strings.TrimSpace(s.Pref.GetString("storeIp", "")) != ""
}

func (s *Session) startPendingDownload() {
	nv := s.Pref.GetString("newAppVer", "")
	if nv == "" || nv == s.Pref.GetString("appVer", Version) {
		return
	}
	if nv == s.Pref.GetString("upgradeFailVer", "") {
		log.Println("upgrade skip: version already failed", nv)
		return
	}
	if !s.ftpConfigured() {
		log.Println("upgrade skip: ftp storeIp not set")
		return
	}
	s.mu.Lock()
	if s.upgrading {
		s.mu.Unlock()
		return
	}
	s.upgrading = true
	s.mu.Unlock()
	s.showUpgradeProgress()
	s.setBusy(true)
	go s.downloadUpgrade(nv)
}

func (s *Session) commitVersion(ver string) {
	s.Pref.PutString("appVer", ver)
	s.Pref.PutString("setAppVer", ver)
	s.Pref.PutString("newAppVer", "")
	s.Pref.PutString("upgradeFailVer", "")
	_ = s.Pref.Save()
	if p, err := upgrade.Current(); err == nil {
		upgrade.ClearStatus(p.Status)
	}
	s.mu.Lock()
	s.message = s.T("toast.upgrade_ok")
	s.mu.Unlock()
	s.refresh()
}

func (s *Session) confirmVersion() {
	ver := s.Pref.GetString("setAppVer", "")
	if ver == "" {
		return
	}
	if s.useREST() {
		s.Pref.PutString("setAppVer", "")
		_ = s.Pref.Save()
		return
	}
	if _, err := s.socket().Put("com", "cV").Put("ver", ver).Request(); err != nil {
		log.Println("cV", err)
		return
	}
	s.Pref.PutString("setAppVer", "")
	_ = s.Pref.Save()
}

func (s *Session) State() UI {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buildUI()
}

func (s *Session) refresh() {
	s.mu.Lock()
	ui := s.buildUI()
	cb := s.OnUI
	s.mu.Unlock()
	if cb != nil {
		cb(ui)
	}
}

func (s *Session) cur() screen {
	if len(s.stack) == 0 {
		return screen{ID: "login", Title: "Prijava"}
	}
	return s.stack[len(s.stack)-1]
}

func (s *Session) push(sc screen) {
	if sc.Args == nil {
		sc.Args = map[string]string{}
	}
	s.stack = append(s.stack, sc)
	s.allowBuy = true
	s.form = map[string]string{}
	s.reportRows = nil
	s.reportLast = nil
}

func (s *Session) Back() {
	s.mu.Lock()
	if len(s.stack) > 1 {
		s.stack = s.stack[:len(s.stack)-1]
	}
	s.confirm = ""
	s.confirmYes = nil
	s.mu.Unlock()
	s.refresh()
}

func (s *Session) Home() {
	s.mu.Lock()
	logged := false
	for _, sc := range s.stack {
		if sc.ID == "home" {
			logged = true
			break
		}
	}
	if logged {
		s.stack = []screen{{ID: "home", Title: "Home"}}
		s.tab = "apps"
	} else {
		s.stack = []screen{{ID: "login", Title: "Prijava"}}
		s.tab = "login"
	}
	s.confirm = ""
	s.confirmYes = nil
	s.mu.Unlock()
	s.refresh()
}

func (s *Session) SetTab(tab string) {
	s.mu.Lock()
	s.tab = tab
	s.mu.Unlock()
	s.refresh()
}

func (s *Session) SetTheme(name string) {
	s.mu.Lock()
	s.theme = name
	s.Pref.PutString("theme-name", name)
	_ = s.Pref.Save()
	label := name
	for _, t := range Themes() {
		if t.ID == name {
			label = t.Label
		}
	}
	s.toast = s.T("toast.theme", label)
	s.mu.Unlock()
	s.refresh()
	s.clearToastLater()
}

func (s *Session) showToast(msg string) {
	s.mu.Lock()
	s.toast = msg
	s.mu.Unlock()
	s.refresh()
	s.clearToastLater()
}

func (s *Session) clearToastLater() {
	go func() {
		time.Sleep(2 * time.Second)
		s.mu.Lock()
		s.toast = ""
		s.mu.Unlock()
		s.refresh()
	}()
}

func (s *Session) setBusy(v bool) {
	s.mu.Lock()
	s.busy = v
	s.mu.Unlock()
	s.refresh()
}

func (s *Session) T(key string, args ...any) string {
	if s.Tr == nil {
		if len(args) == 0 {
			return key
		}
		return key
	}
	return s.Tr.T(key, args...)
}

func (s *Session) Translate(key string) string {
	return s.T(key)
}

func (s *Session) SetLang(code string) {
	code = i18n.Normalize(code)
	s.Tr.SetLang(code)
	s.Pref.PutString("language", code)
	_ = s.Pref.Save()
	label := code
	for _, l := range i18n.Languages() {
		if l.ID == code {
			label = l.Label
		}
	}
	s.showToast(s.T("toast.language", label))
}

func (s *Session) useREST() bool {
	t := strings.ToLower(s.Pref.GetString("transport", "socket"))
	return t == "rest" || t == "http" || t == "fulgor"
}

func (s *Session) rest() *fulgor.Client {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.restClient == nil {
		port := s.Pref.GetInt("serverPort", 8080)
		if port == 0 {
			port = 8080
		}
		s.restClient = fulgor.New(fulgor.Config{
			BaseURL: fulgor.BaseURL(s.Pref.GetString("scheme", "http"), s.Pref.GetString("serverIp", "127.0.0.1"), port),
		})
	}
	return s.restClient
}

func (s *Session) ResetREST() {
	s.mu.Lock()
	c := s.restClient
	s.restClient = nil
	s.mu.Unlock()
	if c != nil {
		_ = c.Logout()
	}
}

func (s *Session) request(h *netproto.Host) (map[string]string, error) {
	if s.useREST() {
		return s.rest().Dispatch(h.Map())
	}
	return h.Request()
}

func (s *Session) socket() *netproto.Host {
	h := netproto.NewHost(s.Pref.GetString("serverIp", ""), s.Pref.GetInt("serverPort", 0))
	h.Put("tsN", s.Pref.GetString("terminalSerialNumber", ""))
	h.Put("pass", s.Mem.GetString("pass", ""))
	return h
}

func (s *Session) nextRC() int {
	next := s.Pref.GetInt("requestCode", 1) + 1
	if next > 999 {
		next = 1
	}
	s.Pref.PutInt("requestCode", next)
	_ = s.Pref.Save()
	return next
}

func (s *Session) printSettings() print.Settings {
	st := s.Pref
	return print.Settings{
		UseFiscal:     st.GetString("useFiscal", "false"),
		UseSerial:     st.GetString("useSerialPort", "false"),
		FiscalType:    st.GetString("fiscalPrinterType", ""),
		SerialType:    st.GetString("serialPrinterType", ""),
		PrinterName:   st.GetString("printerName", ""),
		PrinterFolder: st.GetString("printerFolder", ""),
		FileExt:       st.GetString("fileExt", ""),
		ComPort:       st.GetString("comPort", ""),
		RowLen:        st.GetInt("printerRowLen", 30),
		LeftMargin:    st.GetInt("leftMargin", 0),
		TopMargin:     st.GetInt("topMargin", 0),
		PageWidth:     st.GetInt("pageWidth", 80),
		PageScale:     st.GetFloat("pageScale", 0.8),
	}
}

func (s *Session) printable(service string) bool {
	if service == "true" {
		return true
	}
	return strings.Contains(s.Mem.GetString("printable", ""), service)
}

func (s *Session) doPrint(params map[string]string, tmpl *print.Template, service string, showMsg bool) {
	if tmpl == nil || !s.printable(service) {
		if showMsg && !s.printable(service) {
			s.showToast(s.T("toast.print_off", service))
		}
		return
	}
	if showMsg {
		s.showToast(s.T("toast.printing"))
	}
	job := &print.Job{Settings: s.printSettings(), Template: tmpl, Params: print.FromMap(params)}
	go func() {
		if err := job.Execute(); err != nil {
			if errors.Is(err, print.ErrNoPrinter) {
				s.showToast(s.T("toast.printer_missing"))
				return
			}
			s.showToast(err.Error())
		}
	}()
}

func (s *Session) printAll(params map[string]string, full, short *print.Template, service string) {
	s.doPrint(params, full, service, true)
	s.doPrint(params, short, service, false)
}

func (s *Session) runNet(h *netproto.Host, ok func(map[string]string), fail func(error)) {
	s.setBusy(true)
	go func() {
		m, err := s.request(h)
		s.setBusy(false)
		if err != nil {
			if fail != nil {
				fail(err)
			} else {
				s.showToast(err.Error())
			}
			return
		}
		ok(m)
	}()
}

func (s *Session) doAndWaitCheck(host *netproto.Host, rc int, cmd, check string, then func(map[string]string)) {
	h := host.Clone().Put("rc", strconv.Itoa(rc)).Put("com", cmd)
	s.setBusy(true)
	go func() {
		m, err := s.request(h)
		finish := func(res map[string]string) {
			s.setBusy(false)
			then(res)
		}
		failHome := func(msg string) {
			s.setBusy(false)
			s.showToast(msg)
			s.Home()
		}
		if err != nil {
			if netproto.IsReadTimeout(err) || fulgor.IsTimeout(err) {
				cm, cerr := s.request(host.Clone().Put("rc", strconv.Itoa(rc)).Put("com", check))
				if cerr != nil {
					failHome(cerr.Error())
					return
				}
				finish(cm)
				return
			}
			failHome(err.Error())
			return
		}
		if m["err"] == "210" {
			cm, cerr := s.request(host.Clone().Put("rc", strconv.Itoa(rc)).Put("com", check))
			if cerr != nil {
				failHome(cerr.Error())
				return
			}
			finish(cm)
			return
		}
		finish(m)
	}()
}

func (s *Session) Confirm(ok bool) {
	s.mu.Lock()
	yes := s.confirmYes
	s.confirm = ""
	s.confirmYes = nil
	s.mu.Unlock()
	if ok && yes != nil {
		yes()
		return
	}
	s.refresh()
}

func (s *Session) ask(text string, yes func()) {
	s.mu.Lock()
	s.confirm = strings.TrimSpace(text)
	s.confirmYes = yes
	s.mu.Unlock()
	s.refresh()
}

func (s *Session) Printers() []string { return print.ListPrinters() }
func (s *Session) SerialPorts() []string {
	ports := hardware.ListSerialPorts()
	if len(ports) == 0 {
		if p := s.Pref.GetString("comPort", ""); p != "" {
			return []string{p}
		}
	}
	return ports
}

func (s *Session) startActive() {
	s.stopActive = make(chan struct{})
	go func() {
		t := time.NewTicker(5 * time.Minute)
		defer t.Stop()
		for {
			select {
			case <-s.stopActive:
				return
			case <-t.C:
				s.checkActive()
			}
		}
	}()
}

func (s *Session) checkActive() {
	if s.Mem.GetString("pass", "") == "" {
		return
	}
	s.confirmVersion()
	var m map[string]string
	var err error
	if s.useREST() {
		m, err = s.rest().Login(s.Pref.GetString("terminalSerialNumber", ""))
	} else {
		m, err = s.socket().Put("com", "LogIn").Request()
	}
	if err != nil {
		log.Println("active", err)
		return
	}
	if m["err"] != "60" {
		return
	}
	s.applySession(m)
	s.offerUpgrade()
}

func (s *Session) applySession(m map[string]string) {
	var ov operators.Overlay
	if s.useREST() {
		ov = s.rest().Operators()
	}
	s.Mem.PutString("printable", m["printable"])
	s.mu.Lock()
	s.opCfg = ov
	s.services = m["servisi"]
	filtered := filterMetas(s.apps, s.services)
	if len(filtered) == 1 && s.cur().ID == "home" {
		s.Mem.PutBool("one", true)
		s.stack = []screen{{ID: filtered[0].Activity, Title: filtered[0].Title}}
	}
	s.mu.Unlock()
	s.refresh()
	newVer := m["appVer"]
	if newVer == "" {
		newVer = "1.0"
	}
	cur := s.Pref.GetString("appVer", "1.0")
	if newVer != cur && s.Pref.GetString("setAppVer", "") == "" {
		if skipSelfUpdate() {
			log.Printf("upgrade skip: wails dev server=%s local=%s", newVer, cur)
			return
		}
		if newVer == s.Pref.GetString("upgradeFailVer", "") {
			log.Println("upgrade skip: version already failed", newVer)
			return
		}
		s.Pref.PutString("newAppVer", newVer)
		_ = s.Pref.Save()
	}
}

func (s *Session) downloadUpgrade(ver string) {
	install := s.installUpgrade
	if install == nil {
		install = upgrade.Install
	}
	if err := install(s.Pref, ver, s.lockDir()); err != nil {
		log.Println("upgrade download failed, staying open:", err)
		s.mu.Lock()
		s.upgrading = false
		s.mu.Unlock()
		s.setBusy(false)
		s.showToast(s.T("toast.upgrade_fail", err.Error()))
		return
	}
	log.Println("upgrade download ok, closing to apply", ver)
	s.showUpgradeProgress()
	s.mu.Lock()
	s.message = s.T("toast.upgrade_restart")
	s.mu.Unlock()
	s.refresh()
	time.Sleep(400 * time.Millisecond)
	s.quit()
}

func (s *Session) Login(pass string) {
	s.Mem.PutString("pass", pass)
	s.setBusy(true)
	go func() {
		if !s.registerIfNeeded() {
			s.setBusy(false)
			s.showToast(s.T("toast.register_fail"))
			return
		}
		if s.useREST() {
			m, err := s.rest().Login(s.Pref.GetString("terminalSerialNumber", ""))
			s.setBusy(false)
			if err != nil {
				s.showToast(err.Error())
				return
			}
			if m["err"] != "" && m["err"] != "60" {
				s.showToast(s.T(m["err"]))
				return
			}
			s.applySession(m)
			s.mu.Lock()
			s.stack = []screen{{ID: "home", Title: "Home"}}
			s.tab = "apps"
			s.mu.Unlock()
			s.offerUpgrade()
			s.refresh()
			return
		}
		m, err := s.socket().Put("com", "gV").Put("pass", pass).Request()
		s.setBusy(false)
		if err != nil {
			s.showToast(err.Error())
			return
		}
		if _, has := m["err"]; has {
			s.showToast(s.T(m["err"]))
			return
		}
		go s.checkActive()
		s.mu.Lock()
		s.stack = []screen{{ID: "home", Title: "Home"}}
		s.tab = "apps"
		s.mu.Unlock()
		s.offerUpgrade()
		s.refresh()
	}()
}

func (s *Session) registerIfNeeded() bool {
	if s.Pref.GetBool("registered", false) {
		return true
	}
	var m map[string]string
	var err error
	if s.useREST() {
		m, err = s.rest().Register(s.Pref.GetString("userSerialNumber", ""))
	} else {
		h := netproto.NewHost(s.Pref.GetString("serverIp", ""), s.Pref.GetInt("serverPort", 0))
		h.Put("tsN", s.Pref.GetString("userSerialNumber", ""))
		h.Put("pass", s.Mem.GetString("pass", ""))
		h.Put("com", "rA")
		m, err = h.Request()
	}
	if err != nil {
		s.showToast(err.Error())
		return false
	}
	errCode := m["err"]
	if errCode == "" {
		errCode = "1022"
	}
	if errCode == "116" {
		s.Pref.PutString("terminalSerialNumber", m["nSn"])
		s.Pref.PutBool("registered", true)
		_ = s.Pref.Save()
		return true
	}
	s.showToast(s.T(errCode))
	return false
}

func (s *Session) Key(ch string) {
	rs := []rune(ch)
	if len(rs) == 0 {
		return
	}
	if ev := s.cards.HandleRune(rs[0]); ev != nil {
		s.onCard(ev.CardNumber)
		return
	}
	if !s.Pref.GetBool("userShortCut", false) {
		return
	}
	if rs[0] >= '1' && rs[0] <= '9' {
		s.Digit(int(rs[0] - '0'))
	}
}

func (s *Session) KeyEnter() {
	if ev := s.cards.HandleEnter(); ev != nil {
		s.onCard(ev.CardNumber)
		return
	}
	s.mu.Lock()
	id := s.cur().ID
	s.mu.Unlock()
	if id == "login" {
		s.mu.Lock()
		pass := s.form["pass"]
		s.mu.Unlock()
		s.Login(pass)
		return
	}
	s.Submit(nil)
}

func (s *Session) Digit(n int) {
	s.mu.Lock()
	ui := s.buildUI()
	s.mu.Unlock()
	if n <= 0 || n > len(ui.Items) {
		return
	}
	s.Tap(ui.Items[n-1].ID)
}

func (s *Session) onCard(number string) {
	s.mu.Lock()
	cur := s.cur().ID
	s.mu.Unlock()
	dest, ok := operators.OnCard(cur, number)
	if !ok {
		return
	}
	if dest != "" {
		s.open(dest)
		return
	}
	s.refresh()
}
