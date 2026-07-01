// This file verifies linapro-monitor-server plugin callback wiring helpers.

package backend

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gcfg"

	"lina-core/pkg/plugin/capability/plugincap"
	monitorsvc "lina-plugin-linapro-monitor-server/backend/internal/service/monitor"
)

// fakeMonitorService records callback usage without touching the database or host metrics.
type fakeMonitorService struct {
	// collected reports whether CollectAndStore was called.
	collected bool
	// cleanupCalled reports whether CleanupStale was called.
	cleanupCalled bool
	// cleanupThreshold records the threshold passed to CleanupStale.
	cleanupThreshold time.Duration
}

// CollectAndStore records one collection callback.
func (s *fakeMonitorService) CollectAndStore(ctx context.Context) {
	s.collected = true
}

// Collect satisfies monitorsvc.Service for tests.
func (s *fakeMonitorService) Collect(ctx context.Context) *monitorsvc.MonitorData {
	return nil
}

// GetDBInfo satisfies monitorsvc.Service for tests.
func (s *fakeMonitorService) GetDBInfo(ctx context.Context) *monitorsvc.DBInfo {
	return nil
}

// GetLatest satisfies monitorsvc.Service for tests.
func (s *fakeMonitorService) GetLatest(ctx context.Context, nodeName string) ([]*monitorsvc.NodeMonitorData, error) {
	return nil, nil
}

// CleanupStale records one cleanup callback.
func (s *fakeMonitorService) CleanupStale(ctx context.Context, threshold time.Duration) (int64, error) {
	s.cleanupCalled = true
	s.cleanupThreshold = threshold
	return 0, nil
}

// TestCollectSnapshotUsesInjectedService verifies cron collection reuses the provided service instance.
func TestCollectSnapshotUsesInjectedService(t *testing.T) {
	monitorSvc := &fakeMonitorService{}

	if err := collectSnapshot(context.Background(), monitorSvc); err != nil {
		t.Fatalf("collect snapshot: %v", err)
	}

	if !monitorSvc.collected {
		t.Fatal("expected injected monitor service to collect")
	}
}

// TestCleanupSnapshotsSkipsNonPrimaryNode verifies cleanup is skipped outside the primary node.
func TestCleanupSnapshotsSkipsNonPrimaryNode(t *testing.T) {
	monitorSvc := &fakeMonitorService{}

	if err := cleanupSnapshots(context.Background(), false, nil, monitorSvc); err != nil {
		t.Fatalf("cleanup snapshots: %v", err)
	}

	if monitorSvc.cleanupCalled {
		t.Fatal("expected non-primary node to skip cleanup")
	}
}

// TestCleanupSnapshotsUsesInjectedServiceOnPrimaryNode verifies cleanup uses the shared service instance.
func TestCleanupSnapshotsUsesInjectedServiceOnPrimaryNode(t *testing.T) {
	configSvc := newPluginTestConfigService(t, `
monitor:
  interval: 30s
  retentionMultiplier: 4
`)

	monitorSvc := &fakeMonitorService{}

	if err := cleanupSnapshots(context.Background(), true, configSvc, monitorSvc); err != nil {
		t.Fatalf("cleanup snapshots: %v", err)
	}

	if !monitorSvc.cleanupCalled {
		t.Fatal("expected injected monitor service to clean up")
	}
	if monitorSvc.cleanupThreshold != 2*time.Minute {
		t.Fatalf("expected cleanup threshold 2m, got %s", monitorSvc.cleanupThreshold)
	}
}

// newPluginTestConfigService builds a plugin-domain service from artifact content.
func newPluginTestConfigService(t *testing.T, content string) plugincap.Service {
	t.Helper()

	adapter, err := gcfg.NewAdapterContent(content)
	if err != nil {
		t.Fatalf("create test config adapter: %v", err)
	}
	return pluginTestPlugins{config: pluginTestConfigService{cfg: gcfg.NewWithAdapter(adapter)}}
}

// pluginTestPlugins exposes plugin-domain capabilities used by backend tests.
type pluginTestPlugins struct {
	config plugincap.ConfigService
}

// Config returns the configured test plugin config service.
func (s pluginTestPlugins) Config() plugincap.ConfigService {
	return s.config
}

// Registry is unused by backend tests.
func (s pluginTestPlugins) Registry() plugincap.RegistryService {
	return nil
}

// State is unused by backend tests.
func (s pluginTestPlugins) State() plugincap.StateService {
	return nil
}

// Lifecycle is unused by backend tests.
func (s pluginTestPlugins) Lifecycle() plugincap.LifecycleService {
	return nil
}

// pluginTestConfigService is a test-local plugincap.ConfigService backed by YAML content.
type pluginTestConfigService struct {
	cfg *gcfg.Config
}

// Get returns one raw test config value.
func (s pluginTestConfigService) Get(ctx context.Context, key string, defaultValue any) (*gvar.Var, error) {
	if s.cfg == nil {
		if defaultValue != nil {
			return gvar.New(defaultValue), nil
		}
		return nil, nil
	}
	if defaultValue != nil {
		return s.cfg.Get(ctx, key, defaultValue)
	}
	return s.cfg.Get(ctx, key)
}

// Exists reports whether one test config key exists.
func (s pluginTestConfigService) Exists(ctx context.Context, key string) (bool, error) {
	value, err := s.Get(ctx, key, nil)
	return value != nil && !value.IsNil(), err
}

// Scan scans one test config section into target.
func (s pluginTestConfigService) Scan(ctx context.Context, key string, target any) error {
	if target == nil {
		return gerror.New("plugin config scan target cannot be nil")
	}
	value, err := s.Get(ctx, key, nil)
	if err != nil || value == nil || value.IsNil() {
		return err
	}
	return value.Scan(target)
}

// String reads a string test config value.
func (s pluginTestConfigService) String(ctx context.Context, key string, defaultValue string) (string, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	if raw := value.String(); strings.TrimSpace(raw) != "" {
		return raw, nil
	}
	return defaultValue, nil
}

// Bool reads a bool test config value.
func (s pluginTestConfigService) Bool(ctx context.Context, key string, defaultValue bool) (bool, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.Bool(), nil
}

// Int reads an int test config value.
func (s pluginTestConfigService) Int(ctx context.Context, key string, defaultValue int) (int, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.Int(), nil
}

// Duration reads a duration test config value.
func (s pluginTestConfigService) Duration(ctx context.Context, key string, defaultValue time.Duration) (time.Duration, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	raw := strings.TrimSpace(value.String())
	if raw == "" {
		return defaultValue, nil
	}
	return time.ParseDuration(raw)
}
