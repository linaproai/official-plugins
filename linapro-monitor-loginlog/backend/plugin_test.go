// This file verifies linapro-monitor-loginlog plugin cleanup cron helpers.

package backend

import (
	"context"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	loginlogsvc "lina-plugin-linapro-monitor-loginlog/backend/internal/service/loginlog"
	"testing"

	"github.com/gogf/gf/v2/container/gvar"
)

// loginLogCleanupHostConfigStub returns a deterministic retention value.
type loginLogCleanupHostConfigStub struct {
	hostconfigcap.Service
	value *gvar.Var
	err   error
}

// Get returns the configured retention value.
func (s loginLogCleanupHostConfigStub) Get(context.Context, string, any) (*gvar.Var, error) {
	return s.value, s.err
}

// loginLogCleanupCleanerStub records cleanup calls.
type loginLogCleanupCleanerStub struct {
	loginlogsvc.Service
	called bool
	days   int
	err    error
}

// CleanupExpired records the retention period passed by the cron helper.
func (s *loginLogCleanupCleanerStub) CleanupExpired(_ context.Context, days int) (int, error) {
	s.called = true
	s.days = days
	return 0, s.err
}

// TestCleanupExpiredLoginLogsSkipsNonPrimaryNode verifies non-primary nodes do
// not require dependencies or execute cleanup.
func TestCleanupExpiredLoginLogsSkipsNonPrimaryNode(t *testing.T) {
	cleaner := &loginLogCleanupCleanerStub{}
	if err := cleanupExpiredLoginLogs(context.Background(), false, nil, cleaner); err != nil {
		t.Fatalf("cleanup on non-primary node: %v", err)
	}
	if cleaner.called {
		t.Fatal("expected non-primary node to skip login-log cleanup")
	}
}

// TestCleanupExpiredLoginLogsUsesHostRetention verifies the cron helper reads
// the global retention period from host config.
func TestCleanupExpiredLoginLogsUsesHostRetention(t *testing.T) {
	cleaner := &loginLogCleanupCleanerStub{}
	hostConfig := loginLogCleanupHostConfigStub{value: gvar.New("120")}

	if err := cleanupExpiredLoginLogs(context.Background(), true, hostConfig, cleaner); err != nil {
		t.Fatalf("cleanup login logs: %v", err)
	}
	if !cleaner.called || cleaner.days != 120 {
		t.Fatalf("expected cleanup days=120, got called=%t days=%d", cleaner.called, cleaner.days)
	}
}

// TestCleanupExpiredLoginLogsRejectsInvalidRetention verifies defensive
// validation in case a host-config adapter bypasses protected parameter checks.
func TestCleanupExpiredLoginLogsRejectsInvalidRetention(t *testing.T) {
	cleaner := &loginLogCleanupCleanerStub{}
	hostConfig := loginLogCleanupHostConfigStub{value: gvar.New("0")}

	if err := cleanupExpiredLoginLogs(context.Background(), true, hostConfig, cleaner); err == nil {
		t.Fatal("expected invalid retention days to fail")
	}
	if cleaner.called {
		t.Fatal("expected invalid retention days not to call cleanup")
	}
}

// TestCleanupExpiredLoginLogsRequiresRetention verifies the plugin does not
// synthesize a default when the host runtime parameter is absent.
func TestCleanupExpiredLoginLogsRequiresRetention(t *testing.T) {
	cleaner := &loginLogCleanupCleanerStub{}
	hostConfig := loginLogCleanupHostConfigStub{}

	if err := cleanupExpiredLoginLogs(context.Background(), true, hostConfig, cleaner); err == nil {
		t.Fatal("expected missing retention days to fail")
	}
	if cleaner.called {
		t.Fatal("expected missing retention days not to call cleanup")
	}
}
