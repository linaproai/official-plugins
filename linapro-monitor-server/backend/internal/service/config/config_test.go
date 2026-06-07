// This file verifies linapro-monitor-server plugin configuration loading.

package config

import (
	"context"
	"strings"
	"testing"
	"time"

	"lina-core/pkg/plugin/capability/plugincap"
	configsvc "lina-core/pkg/plugin/capability/plugincap"
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

// newTestConfigService builds a scoped plugin config reader from artifact content.
func newTestConfigService(t *testing.T, content string) plugincap.ConfigService {
	t.Helper()

	return configsvc.NewConfigFactory(t.TempDir(), t.TempDir()).
		WithArtifactConfig("linapro-monitor-server", []byte(content)).
		ForPlugin("linapro-monitor-server")
}
