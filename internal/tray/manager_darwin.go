//go:build darwin

// Package tray provides macOS-specific system tray integration using DarwinKit.
//
// IMPORTANT: This file uses DarwinKit instead of fyne.io/systray to avoid
// AppDelegate symbol conflicts with Wails on macOS.
//
// Build Requirements:
//   - Always use `-tags nosystray` when building on macOS
//   - This prevents fyne.io/systray's Objective-C code from being compiled
//
// The Makefile and GitHub Actions workflow automatically apply the correct tags.
package tray

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/progrium/darwinkit/dispatch"
	"github.com/progrium/darwinkit/helper/action"
	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/macos/foundation"
	"github.com/progrium/darwinkit/objc"

	"MrRSS/internal/handlers/core"
)

// Manager provides a thin wrapper around the macOS status bar menu using DarwinKit.
type Manager struct {
	handler    *core.Handler
	icon       []byte
	running    atomic.Bool
	stopCh     chan struct{}
	mu         sync.Mutex
	lang       string
	statusItem appkit.StatusItem
}

// NewManager creates a new tray manager for macOS.
func NewManager(handler *core.Handler, icon []byte) *Manager {
	return &Manager{
		handler: handler,
		icon:    icon,
	}
}

// Start initialises the macOS status bar menu if it isn't already running.
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

	// Run on main thread as required by AppKit
	dispatch.MainQueue().DispatchAsync(func() {
		m.setupStatusBar(ctx, onQuit, onShow)
	})

	// Watch for stop signal
	go func() {
		select {
		case <-ctx.Done():
			m.cleanup()
		case <-m.stopCh:
			m.cleanup()
		}
	}()
}

func (m *Manager) setupStatusBar(ctx context.Context, onQuit func(), onShow func()) {
	labels := m.getLabels()

	// Get the system status bar
	statusBar := appkit.StatusBar_SystemStatusBar()

	// Create a status item with variable length
	m.statusItem = statusBar.StatusItemWithLength(appkit.VariableStatusItemLength)
	objc.Retain(m.statusItem)

	// Set up the button
	button := m.statusItem.Button()
	if button.Ptr() != nil {
		button.SetTitle(labels.title)

		// Try to set image from icon bytes
		if len(m.icon) > 0 {
			image := appkit.NewImageWithData(m.icon)
			if image.Ptr() != nil {
				// Scale image to appropriate size for menu bar (16x16 or 18x18)
				image.SetSize(foundation.Size{Width: 18, Height: 18})
				image.SetTemplate(true) // Makes it adapt to light/dark mode
				button.SetImage(image)
				button.SetTitle("") // Clear title when we have an image
			}
		}
	}

	// Create menu
	menu := appkit.NewMenu()
	objc.Retain(menu)

	// Show item
	showItem := appkit.NewMenuItemWithTitleActionKeyEquivalent(labels.show, objc.Sel(""), "")
	showItem.SetTarget(showItem)
	m.setMenuItemAction(showItem, func() {
		if onShow != nil {
			onShow()
		}
	})
	menu.AddItem(showItem)

	// Refresh item
	refreshItem := appkit.NewMenuItemWithTitleActionKeyEquivalent(labels.refresh, objc.Sel(""), "")
	refreshItem.SetTarget(refreshItem)
	m.setMenuItemAction(refreshItem, func() {
		if m.handler != nil && m.handler.Fetcher != nil {
			go m.handler.Fetcher.FetchAll(ctx)
		}
	})
	menu.AddItem(refreshItem)

	// Separator
	menu.AddItem(appkit.MenuItem_SeparatorItem())

	// Quit item
	quitItem := appkit.NewMenuItemWithTitleActionKeyEquivalent(labels.quit, objc.Sel(""), "q")
	quitItem.SetTarget(quitItem)
	m.setMenuItemAction(quitItem, func() {
		if onQuit != nil {
			onQuit()
		}
	})
	menu.AddItem(quitItem)

	m.statusItem.SetMenu(menu)
}

// setMenuItemAction sets up an action handler for a menu item using a closure wrapper
func (m *Manager) setMenuItemAction(item appkit.MenuItem, handler func()) {
	action.Set(item, func(obj objc.Object) {
		handler()
	})
}

func (m *Manager) cleanup() {
	dispatch.MainQueue().DispatchAsync(func() {
		if m.statusItem.Ptr() != nil {
			statusBar := appkit.StatusBar_SystemStatusBar()
			statusBar.RemoveStatusItem(m.statusItem)
			m.statusItem.Release()
		}
		m.running.Store(false)
	})
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

// Stop tears down the status bar item if it is running.
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
	m.cleanup()
}

// IsRunning returns true if the tray has been started.
func (m *Manager) IsRunning() bool {
	return m.running.Load()
}
