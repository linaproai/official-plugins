// This file verifies linapro-monitor-server plugin callback wiring helpers.

package backend

import (
	"context"
	"testing"
	"time"

	pluginconfig "lina-core/pkg/plugin/capability/config"
	plugincontract "lina-core/pkg/plugin/capability/contract"
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

// newPluginTestConfigService builds a scoped plugin config service from artifact content.
func newPluginTestConfigService(t *testing.T, content string) plugincontract.ConfigService {
	t.Helper()

	return pluginconfig.NewFactory(t.TempDir(), t.TempDir()).
		WithArtifactConfig("linapro-monitor-server", []byte(content)).
		ForPlugin("linapro-monitor-server")
}
