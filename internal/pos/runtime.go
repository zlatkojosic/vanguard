package pos

import (
	"nts/vanguard/internal/netproto"
	"nts/vanguard/internal/operators"
	"nts/vanguard/internal/print"
)

var _ operators.Runtime = (*Session)(nil)

func (s *Session) Socket() *netproto.Host { return s.socket() }
func (s *Session) NextRC() int            { return s.nextRC() }
func (s *Session) Pass() string {
	return s.Mem.GetString("pass", "")
}
func (s *Session) Toast(msg string) { s.showToast(msg) }
func (s *Session) Ask(text string, yes func()) {
	s.ask(text, yes)
}
func (s *Session) Print(params map[string]string, tmpl *print.Template, service string, showMsg bool) {
	s.doPrint(params, tmpl, service, showMsg)
}
func (s *Session) PrintAll(params map[string]string, full, short *print.Template, service string) {
	s.printAll(params, full, short, service)
}
func (s *Session) RunNet(h *netproto.Host, ok func(map[string]string), fail func(error)) {
	s.runNet(h, ok, fail)
}
func (s *Session) DoAndWaitCheck(host *netproto.Host, rc int, cmd, check string, then func(map[string]string)) {
	s.doAndWaitCheck(host, rc, cmd, check, then)
}
func (s *Session) Refresh() { s.refresh() }
func (s *Session) Push(id string) {
	s.open(id)
}
func (s *Session) ScreenID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cur().ID
}
func (s *Session) AllowBuy() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.allowBuy
}
func (s *Session) SetAllowBuy(v bool) {
	s.mu.Lock()
	s.allowBuy = v
	s.mu.Unlock()
}
