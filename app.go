package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"nts/vanguard/internal/hardware"
	"nts/vanguard/internal/pos"
	"nts/vanguard/internal/store"
)

type App struct {
	ctx  context.Context
	sess *pos.Session
	lock *hardware.Lock
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	cfg, _ := os.UserConfigDir()
	lockDir := filepath.Join(cfg, "vanguard")
	lock, err := hardware.AcquireLock(lockDir)
	if err != nil {
		_, _ = runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "N POS Z",
			Message: err.Error(),
		})
		runtime.Quit(ctx)
		return
	}
	a.lock = lock

	pref, err := store.NewPref(store.DefaultPrefPath())
	if err != nil {
		runtime.Quit(ctx)
		return
	}
	pos.LoadProperties(pref)
	a.sess = pos.New(pref)
	a.sess.LockDir = lockDir
	a.sess.Quit = func() { runtime.Quit(ctx) }
	title := ""
	a.sess.OnUI = func(u pos.UI) {
		if u.AppTitle != "" && u.AppTitle != title {
			title = u.AppTitle
			runtime.WindowSetTitle(ctx, title)
		}
		runtime.EventsEmit(ctx, "ui", u)
	}
	a.sess.Start()
}

func (a *App) shutdown(ctx context.Context) {
	if a.sess != nil {
		a.sess.ResetREST()
	}
	if a.lock != nil {
		a.lock.Release()
	}
}

func (a *App) State() pos.UI {
	if a.sess == nil {
		return pos.UI{}
	}
	return a.sess.State()
}

func (a *App) Login(pass string) {
	if a.sess != nil {
		a.sess.Login(pass)
	}
}

func (a *App) Back() {
	if a.sess != nil {
		a.sess.Back()
	}
}

func (a *App) Home() {
	if a.sess != nil {
		a.sess.Home()
	}
}

func (a *App) SetTab(tab string) {
	if a.sess != nil {
		a.sess.SetTab(tab)
	}
}

func (a *App) Tap(id string) {
	if a.sess != nil {
		a.sess.Tap(id)
	}
}

func (a *App) Submit(values map[string]string) {
	if a.sess != nil {
		a.sess.Submit(values)
	}
}

func (a *App) Confirm(ok bool) {
	if a.sess != nil {
		a.sess.Confirm(ok)
	}
}

func (a *App) Digit(n int) {
	if a.sess != nil {
		a.sess.Digit(n)
	}
}

func (a *App) Key(ch string) {
	if a.sess != nil {
		a.sess.Key(ch)
	}
}

func (a *App) KeyEnter() {
	if a.sess != nil {
		a.sess.KeyEnter()
	}
}

func (a *App) SetTheme(name string) {
	if a.sess != nil {
		a.sess.SetTheme(name)
	}
}
