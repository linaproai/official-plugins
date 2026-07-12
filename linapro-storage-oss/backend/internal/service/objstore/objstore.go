// Package objstore implements the storagecap.Provider for linapro-storage-oss.
package objstore

import (
	"context"
	"sync"

	"lina-core/pkg/plugin/capability/storagecap"
	settingssvc "lina-plugin-linapro-storage-oss/backend/internal/service/settings"
)

// ConnectionTester probes bucket connectivity.
type ConnectionTester interface {
	TestConnection(ctx context.Context, snapshot *settingssvc.Snapshot) error
}

// Runtime holds the settings service used by the provider factory after route wiring.
type Runtime struct {
	mu       sync.RWMutex
	settings settingssvc.Service
}

// Global is process-local settings injection for the provider factory.
var Global = &Runtime{}

// Configure injects the settings service created during HTTP route registration.
func (r *Runtime) Configure(settings settingssvc.Service) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.settings = settings
}

// Settings returns the configured settings service.
func (r *Runtime) Settings() settingssvc.Service {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.settings
}

// Factory builds a storagecap.Provider from current settings.
func Factory(ctx context.Context, _ storagecap.ProviderEnv) (storagecap.Provider, error) {
	settings := Global.Settings()
	if settings == nil {
		return nil, settingssvc.ErrUnavailable()
	}
	snapshot, err := settings.Load(ctx)
	if err != nil {
		return nil, err
	}
	if err := settings.ValidateReady(snapshot); err != nil {
		return nil, err
	}
	return NewProvider(snapshot)
}

// Tester implements ConnectionTester using the same client construction path.
type Tester struct{}

// TestConnection verifies bucket access.
func (Tester) TestConnection(ctx context.Context, snapshot *settingssvc.Snapshot) error {
	return Probe(ctx, snapshot)
}

// Ensure interfaces are used.
var (
	_ ConnectionTester    = Tester{}
	_ storagecap.Provider = (*provider)(nil)
)
