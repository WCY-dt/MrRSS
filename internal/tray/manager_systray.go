//go:build windows || linux || darwin

package tray

import (
	"context"
	"sync"
	"sync/atomic"

	"fyne.io/systray"

	"MrRSS/internal/handlers/core"
)

// Manager provides a thin wrapper around the system tray menu.
type Manager struct {
	handler *core.Handler
	icon    []byte
	running atomic.Bool
	stopCh  chan struct{}
	mu      sync.Mutex

	lang string
}

// NewManager creates a new tray manager for supported platforms.
func NewManager(handler *core.Handler, icon []byte) *Manager {
	return &Manager{
		handler: handler,
		icon:    icon,
	}
}

// Start initialises the system tray if it isn't already running.
// onQuit should trigger an application shutdown, and onShow should restore the main window.
func (m *Manager) Start(ctx context.Context, onQuit func(), onShow func()) {
	m.mu.Lock()
	if m.running.Load() {
		m.mu.Unlock()
		return
	}

	if m.stopCh != nil {
		select {
		case <-m.stopCh:
		default:
			close(m.stopCh)
		}
		m.stopCh = nil
	}

	m.stopCh = make(chan struct{})
	m.running.Store(true)
	m.mu.Unlock()

	if m.handler != nil && m.handler.DB != nil {
		if lang, err := m.handler.DB.GetSetting("language"); err == nil && lang != "" {
			m.mu.Lock()
			m.lang = lang
			m.mu.Unlock()
		}
	}

	go systray.Run(func() {
		m.run(ctx, onQuit, onShow)
	}, func() {
		m.running.Store(false)
	})
}

func (m *Manager) run(ctx context.Context, onQuit func(), onShow func()) {
	if len(m.icon) > 0 {
		systray.SetIcon(m.icon)
	}
	labels := m.getLabels()

	systray.SetTitle(labels.title)
	systray.SetTooltip(labels.tooltip)

	showItem := systray.AddMenuItem(labels.show, labels.tooltip)
	refreshItem := systray.AddMenuItem(labels.refresh, labels.refreshTooltip)
	systray.AddSeparator()
	quitItem := systray.AddMenuItem(labels.quit, labels.quitTooltip)

	go func() {
		for {
			select {
			case <-showItem.ClickedCh:
				if onShow != nil {
					onShow()
				}
			case <-refreshItem.ClickedCh:
				if m.handler != nil && m.handler.Fetcher != nil {
					go m.handler.Fetcher.FetchAll(ctx)
				}
			case <-quitItem.ClickedCh:
				if onQuit != nil {
					onQuit()
				}
			case <-ctx.Done():
				systray.Quit()
				return
			case <-m.stopCh:
				systray.Quit()
				return
			}
		}
	}()
}

type trayLabels struct {
	title          string
	tooltip        string
	show           string
	refresh        string
	refreshTooltip string
	quit           string
	quitTooltip    string
}

func (m *Manager) getLabels() trayLabels {
	m.mu.Lock()
	lang := m.lang
	m.mu.Unlock()
	switch lang {
	case "zh-CN", "zh", "zh-cn":
		return trayLabels{
			title:          "MrRSS",
			tooltip:        "MrRSS",
			show:           "显示 MrRSS",
			refresh:        "立即刷新",
			refreshTooltip: "刷新所有订阅",
			quit:           "退出",
			quitTooltip:    "退出 MrRSS",
		}
	default:
		return trayLabels{
			title:          "MrRSS",
			tooltip:        "MrRSS",
			show:           "Show MrRSS",
			refresh:        "Refresh now",
			refreshTooltip: "Refresh all feeds",
			quit:           "Quit",
			quitTooltip:    "Quit MrRSS",
		}
	}
}

// Stop tears down the system tray if it is running.
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running.Load() {
		return
	}
	if m.stopCh != nil {
		select {
		case <-m.stopCh:
		default:
			close(m.stopCh)
		}
		m.stopCh = nil
	}
	m.running.Store(false)
}

// IsRunning returns true if the tray has been started.
func (m *Manager) IsRunning() bool {
	return m.running.Load()
}
