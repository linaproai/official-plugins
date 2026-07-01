// This file verifies linapro-monitor-server plugin configuration loading.

package config

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gcfg"

	"lina-core/pkg/plugin/capability/plugincap"
)

// TestLoadUsesDefaultsWhenUnset verifies monitor config defaults when config is absent.
func TestLoadUsesDefaultsWhenUnset(t *testing.T) {
	cfg, err := Load(context.Background(), newTestConfigService(t, `
server:
  address: ":9120"
`))
	if err != nil {
		t.Fatalf("load monitor config: %v", err)
	}
	if cfg.Interval != time.Minute {
		t.Fatalf("expected default interval 1m, got %s", cfg.Interval)
	}
	if cfg.RetentionMultiplier != 5 {
		t.Fatalf("expected default retention multiplier 5, got %d", cfg.RetentionMultiplier)
	}
}

// TestLoadUsesConfiguredValues verifies configured monitor values override defaults.
func TestLoadUsesConfiguredValues(t *testing.T) {
	cfg, err := Load(context.Background(), newTestConfigService(t, `
monitor:
  interval: 45s
  retentionMultiplier: 8
`))
	if err != nil {
		t.Fatalf("load monitor config: %v", err)
	}
	if cfg.Interval != 45*time.Second {
		t.Fatalf("expected interval 45s, got %s", cfg.Interval)
	}
	if cfg.RetentionMultiplier != 8 {
		t.Fatalf("expected retention multiplier 8, got %d", cfg.RetentionMultiplier)
	}
}

// TestLoadReturnsErrorForInvalidDuration verifies invalid duration strings fail config loading.
func TestLoadReturnsErrorForInvalidDuration(t *testing.T) {
	_, err := Load(context.Background(), newTestConfigService(t, `
monitor:
  interval: invalid
`))
	if err == nil {
		t.Fatal("expected invalid duration error")
	}
	if !strings.Contains(err.Error(), "monitor.interval") {
		t.Fatalf("expected error to mention monitor.interval, got %v", err)
	}
}

// TestLoadRejectsSubSecondInterval verifies monitor interval lower bound validation.
func TestLoadRejectsSubSecondInterval(t *testing.T) {
	_, err := Load(context.Background(), newTestConfigService(t, `
monitor:
  interval: 500ms
`))
	if err == nil {
		t.Fatal("expected sub-second interval error")
	}
	if !strings.Contains(err.Error(), "at least 1s") {
		t.Fatalf("expected at least 1s error, got %v", err)
	}
}

// TestLoadRejectsFractionalSecondInterval verifies monitor interval alignment validation.
func TestLoadRejectsFractionalSecondInterval(t *testing.T) {
	_, err := Load(context.Background(), newTestConfigService(t, `
monitor:
  interval: 1500ms
`))
	if err == nil {
		t.Fatal("expected fractional-second interval error")
	}
	if !strings.Contains(err.Error(), "whole seconds") {
		t.Fatalf("expected whole seconds error, got %v", err)
	}
}

// newTestConfigService builds a plugin-domain service from artifact content.
func newTestConfigService(t *testing.T, content string) plugincap.Service {
	t.Helper()

	adapter, err := gcfg.NewAdapterContent(content)
	if err != nil {
		t.Fatalf("create test config adapter: %v", err)
	}
	return testPlugins{config: testConfigService{cfg: gcfg.NewWithAdapter(adapter)}}
}

// testPlugins exposes plugin-domain capabilities used by config tests.
type testPlugins struct {
	config plugincap.ConfigService
}

// Config returns the configured test plugin config service.
func (s testPlugins) Config() plugincap.ConfigService {
	return s.config
}

// Registry is unused by config tests.
func (s testPlugins) Registry() plugincap.RegistryService {
	return nil
}

// State is unused by config tests.
func (s testPlugins) State() plugincap.StateService {
	return nil
}

// Lifecycle is unused by config tests.
func (s testPlugins) Lifecycle() plugincap.LifecycleService {
	return nil
}

// testConfigService is a test-local plugincap.ConfigService backed by YAML content.
type testConfigService struct {
	cfg *gcfg.Config
}

// Get returns one raw test config value.
func (s testConfigService) Get(ctx context.Context, key string, defaultValue any) (*gvar.Var, error) {
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
func (s testConfigService) Exists(ctx context.Context, key string) (bool, error) {
	value, err := s.Get(ctx, key, nil)
	return value != nil && !value.IsNil(), err
}

// Scan scans one test config section into target.
func (s testConfigService) Scan(ctx context.Context, key string, target any) error {
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
func (s testConfigService) String(ctx context.Context, key string, defaultValue string) (string, error) {
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
func (s testConfigService) Bool(ctx context.Context, key string, defaultValue bool) (bool, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.Bool(), nil
}

// Int reads an int test config value.
func (s testConfigService) Int(ctx context.Context, key string, defaultValue int) (int, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	return value.Int(), nil
}

// Duration reads a duration test config value.
func (s testConfigService) Duration(ctx context.Context, key string, defaultValue time.Duration) (time.Duration, error) {
	value, err := s.Get(ctx, key, defaultValue)
	if err != nil || value == nil || value.IsNil() {
		return defaultValue, err
	}
	raw := strings.TrimSpace(value.String())
	if raw == "" {
		return defaultValue, nil
	}
	duration, err := time.ParseDuration(raw)
	if err != nil {
		return 0, gerror.Wrapf(err, "read monitor server interval config failed key=%s", key)
	}
	return duration, nil
}
