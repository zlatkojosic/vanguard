//go:build fyne

package services

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"nts/vanguard/internal/app"
	"nts/vanguard/internal/operators"
	"nts/vanguard/internal/print"
	"nts/vanguard/internal/upgrade"
)

func Register(reg *app.Registry) {
	reg.RegisterService("apps", func() app.Service { return &Apps{} })
	reg.RegisterService("active", func() app.Service { return &Active{} })
	reg.RegisterService("printer", func() app.Service { return &Printer{} })
	reg.RegisterService("card", func() app.Service { return &Card{} })
	reg.RegisterService("upgrade", func() app.Service { return &Upgrade{} })
	reg.RegisterService("aircache", func() app.Service { return &Noop{N: "aircache"} })
}

type Noop struct{ N string }

func (n *Noop) Name() string                  { return n.N }
func (n *Noop) Start(app.Context, app.Intent) {}
func (n *Noop) Stop()                         {}
func (n *Noop) Receive(app.Intent)            {}

type Apps struct {
	ctx app.Context
	svc string
}

func (a *Apps) Name() string { return "apps" }
func (a *Apps) Stop()        {}
func (a *Apps) Start(ctx app.Context, intent app.Intent) {
	a.ctx = ctx
}
func (a *Apps) Receive(in app.Intent) {
	if a.ctx == nil {
		return
	}
	switch in.Action {
	case "ALLOWED_APPS":
		a.svc = in.GetString("services", "")
		a.broadcast()
	case "GET_APPS":
		a.broadcast()
	}
}
func (a *Apps) broadcast() {
	out := app.NewIntent("")
	out.Action = "LOAD_APPLICATION"
	out.PutString("services", a.svc)
	var metas []app.AppMeta
	for _, m := range operators.Metas() {
		metas = append(metas, app.AppMeta{Title: m.Title, Activity: m.Activity, Tags: m.Tags, Always: m.Always})
	}
	out.PutValue("apps", metas)
	a.ctx.SendBroadcast(out)
}

type Active struct {
	ctx  app.Context
	stop chan struct{}
}

func (a *Active) Name() string { return "active" }
func (a *Active) Start(ctx app.Context, intent app.Intent) {
	a.ctx = ctx
	a.stop = make(chan struct{})
	go func() {
		a.check()
		t := time.NewTicker(5 * time.Minute)
		defer t.Stop()
		for {
			select {
			case <-a.stop:
				return
			case <-t.C:
				a.check()
			}
		}
	}()
}
func (a *Active) Stop() {
	if a.stop != nil {
		close(a.stop)
	}
}
func (a *Active) Receive(in app.Intent) {
	if in.Action == "CHECK_APPS" {
		go a.check()
	}
}
func (a *Active) check() {
	if a.ctx == nil {
		return
	}
	h := a.ctx.Socket().Put("com", "LogIn")
	m, err := h.Request()
	if err != nil {
		log.Println("active", err)
		return
	}
	if m["err"] != "60" {
		return
	}
	if ver := a.ctx.Store().GetString("setAppVer", ""); ver != "" {
		if _, err := a.ctx.Socket().Put("com", "cV").Put("ver", ver).Request(); err != nil {
			log.Println("cV", err)
		} else {
			a.ctx.Store().PutString("setAppVer", "")
			_ = a.ctx.Store().Save()
		}
	}
	services := m["servisi"]
	printable := m["printable"]
	a.ctx.Memory().PutString("printable", printable)
	out := app.NewIntent("")
	out.Action = "ALLOWED_APPS"
	out.PutString("services", services)
	out.PutString("printable", printable)
	app.Do(func() { a.ctx.SendBroadcast(out) })
	newVer := m["appVer"]
	if newVer == "" {
		newVer = "1.0"
	}
	cur := a.ctx.Store().GetString("appVer", "1.0")
	if newVer != cur && a.ctx.Store().GetString("setAppVer", "") == "" {
		a.ctx.Store().PutString("newAppVer", newVer)
		_ = a.ctx.Store().Save()
		a.ctx.ShowMessage("Postoji nova verzija.Uradite restart!")
	}
}

type Printer struct{ ctx app.Context }

func (p *Printer) Name() string                        { return "printer" }
func (p *Printer) Start(ctx app.Context, _ app.Intent) { p.ctx = ctx }
func (p *Printer) Stop()                               {}
func (p *Printer) Receive(in app.Intent) {
	if in.Action != "PRINT" || p.ctx == nil {
		return
	}
	v, ok := in.GetValue("printContext")
	if !ok {
		return
	}
	job, ok := v.(*print.Job)
	if !ok {
		return
	}
	go func() {
		if err := job.Execute(); err != nil {
			p.ctx.ShowToast(err.Error())
		}
	}()
}

type Card struct {
	ctx  app.Context
	apps map[string]string
}

func (c *Card) Name() string { return "card" }
func (c *Card) Start(ctx app.Context, _ app.Intent) {
	c.ctx = ctx
	c.apps = map[string]string{}
}
func (c *Card) Stop() {}
func (c *Card) Receive(in app.Intent) {
	if c.ctx == nil {
		return
	}
	switch in.Action {
	case "REGISTER_CARD_APP":
		c.apps[in.GetString("startCardNumber", "")] = in.GetString("appName", "")
	case "SHOW_CARD_APP":
		num := in.GetString("cardNumber", "")
		if len(num) < 3 {
			return
		}
		name := c.apps[num[:3]]
		if name == "" {
			return
		}
		out := app.NewIntent(name)
		out.Action = "CARD_READ"
		out.PutString("cardNumber", num)
		if !c.ctx.SendToCurrent(out) {
			c.ctx.StartActivity(out)
		}
	}
}

type Upgrade struct {
	ctx app.Context
}

func (u *Upgrade) Name() string                        { return "upgrade" }
func (u *Upgrade) Start(ctx app.Context, _ app.Intent) { u.ctx = ctx }
func (u *Upgrade) Stop()                               {}
func (u *Upgrade) Receive(in app.Intent) {
	if in.Action != "DOWNLOAD_APP" || u.ctx == nil {
		return
	}
	ver := in.GetString("newAppVer", "1.0")
	go func() {
		cfg, _ := os.UserConfigDir()
		lockDir := filepath.Join(cfg, "vanguard")
		if err := upgrade.Install(u.ctx.Store(), ver, lockDir); err != nil {
			u.ctx.ShowToast("upgrade nije uspješan: " + err.Error())
			return
		}
		u.ctx.ShowMessage("Aplikacija se zatvara da bi se instalirala nova verzija.")
		time.Sleep(400 * time.Millisecond)
		os.Exit(0)
	}()
}
