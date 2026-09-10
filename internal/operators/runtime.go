package operators

import (
	"nts/vanguard/internal/netproto"
	"nts/vanguard/internal/print"
)

type Text func(key string, args ...any) string

func Keys(key string, _ ...any) string { return key }

type Runtime interface {
	Socket() *netproto.Host
	NextRC() int
	Pass() string
	T(key string, args ...any) string
	Print(params map[string]string, tmpl *print.Template, service string, showMsg bool)
	PrintAll(params map[string]string, full, short *print.Template, service string)
	RunNet(h *netproto.Host, ok func(map[string]string), fail func(error))
	DoAndWaitCheck(host *netproto.Host, rc int, cmd, check string, then func(map[string]string))
	Toast(msg string)
	Ask(text string, yes func())
	Home()
	Push(id string)
	Refresh()
	AllowBuy() bool
	SetAllowBuy(v bool)
	ScreenID() string
}

type Handler interface {
	Owns(id string) bool
	View(id string, t Text) Screen
	Submit(id string, form map[string]string, rt Runtime)
	Tap(id string, act Action, rt Runtime) bool
}

var (
	lamp     = &lampicaOp{}
	handlers = []Handler{
		lamp,
		aircacheOp{},
		saleOp{},
		xbonSMSOp{},
		binbinOp{},
		genericOp{},
	}
)

func Lookup(id string) Handler {
	for _, h := range handlers {
		if h.Owns(id) {
			return h
		}
	}
	return nil
}

func View(id string, t Text) Screen {
	if t == nil {
		t = Keys
	}
	if h := Lookup(id); h != nil {
		return h.View(id, t)
	}
	return infoScreen(id)
}

func Submit(id string, form map[string]string, rt Runtime) {
	if h := Lookup(id); h != nil {
		h.Submit(id, form, rt)
	}
}

func Tap(screenID string, act Action, rt Runtime) bool {
	if h := Lookup(screenID); h != nil {
		return h.Tap(screenID, act, rt)
	}
	return false
}

func OnCard(current, number string) (navigate string, ok bool) {
	return lamp.OnCard(current, number)
}

func confirmText(f *Flow, rt Runtime) string {
	if f.Dialog != "" {
		return rt.T(f.Dialog)
	}
	return rt.T("dialog.flow", f.OperatorName)
}
