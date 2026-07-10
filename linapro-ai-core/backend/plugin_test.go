// This file verifies linapro-ai-core plugin cleanup cron helpers.

package backend

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/container/gvar"

	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/pluginhost"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/spi"
	aisvc "lina-plugin-linapro-ai-core/backend/internal/service/ai"
)

// invocationCleanupHostConfigStub returns a deterministic retention value.
type invocationCleanupHostConfigStub struct {
	hostconfigcap.Service
	value *gvar.Var
	err   error
}

// Get returns the configured retention value.
func (s invocationCleanupHostConfigStub) Get(context.Context, string, any) (*gvar.Var, error) {
	return s.value, s.err
}

// invocationCleanupCleanerStub records cleanup calls.
type invocationCleanupCleanerStub struct {
	aisvc.Service
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

func TestSourcePluginRegistersOwnerAIDescriptor(t *testing.T) {
	definition, ok := pluginhost.GetSourcePlugin(pluginID)
	if !ok {
		t.Fatalf("expected source plugin %s to be registered", pluginID)
	}
	descriptors := definition.GetCapabilityDescriptors()
	if len(descriptors) != 1 {
		t.Fatalf("expected one AI owner descriptor, got %#v", descriptors)
	}
	descriptor := descriptors[0]
	if descriptor.OwnerPluginID != spi.OwnerPluginID ||
		descriptor.Service != spi.ServiceAI ||
		descriptor.Version != spi.VersionV1 ||
		descriptor.SourceContract != spi.SourceContract ||
		descriptor.DynamicContract != spi.DynamicContract {
		t.Fatalf("unexpected AI owner descriptor: %#v", descriptor)
	}
	if descriptor.Invoker == nil {
		t.Fatal("registered AI owner descriptor must retain runtime invoker")
	}

	methods := make(map[string]struct{}, len(descriptor.Methods))
	for _, method := range descriptor.Methods {
		methods[method.Method] = struct{}{}
	}
	for _, method := range []string{
		spi.MethodTextGenerate,
		spi.MethodTextStatusGet,
		spi.MethodStatusesBatchGet,
	} {
		if _, ok := methods[method]; !ok {
			t.Fatalf("expected registered AI descriptor to publish %s", method)
		}
	}
	for _, method := range []string{
		spi.MethodImageGenerate,
		spi.MethodVideoOperationCancel,
	} {
		if _, ok := methods[method]; ok {
			t.Fatalf("unpublished multimodal method %s must not be registered yet", method)
		}
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
