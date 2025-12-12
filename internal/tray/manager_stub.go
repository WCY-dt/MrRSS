//go:build nosystray

package tray

import (
	"context"

	"MrRSS/internal/handlers/core"
)

// Manager is a no-op implementation for unsupported platforms.
type Manager struct{}

func NewManager(_ *core.Handler, _ []byte) *Manager { return &Manager{} }

func (m *Manager) Start(_ context.Context, _ func(), _ func()) {}

func (m *Manager) Stop() {}

func (m *Manager) IsRunning() bool { return false }
