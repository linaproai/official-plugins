// This file verifies linapro-ai-core plugin cleanup cron helpers.

package backend

import (
	"context"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"testing"

	"github.com/gogf/gf/v2/container/gvar"
)

// invocationCleanupHostConfigStub returns a deterministic retention value.
type invocationCleanupHostConfigStub struct {
	hostconfigcap.Service
	value *gvar.Var
	err   error
}

// Get returns the configured retention value.
func (s invocationCleanupHostConfigStub) Get(context.Context, string) (*gvar.Var, error) {
	return s.value, s.err
}

// invocationCleanupCleanerStub records cleanup calls.
type invocationCleanupCleanerStub struct {
	called bool
	days   int
	err    error
}

// CleanupExpiredInvocations records the retention period passed by the cron helper.
func (s *invocationCleanupCleanerStub) CleanupExpiredInvocations(_ context.Context, days int) (int, error) {
	s.called = true
	s.days = days
	return 0, s.err
}

// TestCleanupExpiredInvocationsSkipsNonPrimaryNode verifies non-primary nodes
// do not require dependencies or execute cleanup.
func TestCleanupExpiredInvocationsSkipsNonPrimaryNode(t *testing.T) {
	cleaner := &invocationCleanupCleanerStub{}
	if err := cleanupExpiredInvocations(context.Background(), false, nil, cleaner); err != nil {
		t.Fatalf("cleanup on non-primary node: %v", err)
	}
	if cleaner.called {
		t.Fatal("expected non-primary node to skip invocation-log cleanup")
	}
}

// TestCleanupExpiredInvocationsUsesHostRetention verifies the cron helper reads
// the global retention period from host config.
func TestCleanupExpiredInvocationsUsesHostRetention(t *testing.T) {
	cleaner := &invocationCleanupCleanerStub{}
	hostConfig := invocationCleanupHostConfigStub{value: gvar.New("120")}

	if err := cleanupExpiredInvocations(context.Background(), true, hostConfig, cleaner); err != nil {
		t.Fatalf("cleanup invocation logs: %v", err)
	}
	if !cleaner.called || cleaner.days != 120 {
		t.Fatalf("expected cleanup days=120, got called=%t days=%d", cleaner.called, cleaner.days)
	}
}

// TestCleanupExpiredInvocationsRejectsInvalidRetention verifies defensive
// validation in case a host-config adapter bypasses protected parameter checks.
func TestCleanupExpiredInvocationsRejectsInvalidRetention(t *testing.T) {
	cleaner := &invocationCleanupCleanerStub{}
	hostConfig := invocationCleanupHostConfigStub{value: gvar.New("0")}

	if err := cleanupExpiredInvocations(context.Background(), true, hostConfig, cleaner); err == nil {
		t.Fatal("expected invalid retention days to fail")
	}
	if cleaner.called {
		t.Fatal("expected invalid retention days not to call cleanup")
	}
}

// TestCleanupExpiredInvocationsRequiresRetention verifies the plugin does not
// synthesize a default when the host runtime parameter is absent.
func TestCleanupExpiredInvocationsRequiresRetention(t *testing.T) {
	cleaner := &invocationCleanupCleanerStub{}
	hostConfig := invocationCleanupHostConfigStub{}

	if err := cleanupExpiredInvocations(context.Background(), true, hostConfig, cleaner); err == nil {
		t.Fatal("expected missing retention days to fail")
	}
	if cleaner.called {
		t.Fatal("expected missing retention days not to call cleanup")
	}
}
