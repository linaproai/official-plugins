// This file verifies host-call demo helpers that do not require a running Wasm
// host service.

package dynamicservice

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/container/gvar"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/tenantcap"
)

// fakePluginConfigService returns deterministic plugin config values for unit
// tests.
type fakePluginConfigService struct {
	strings map[string]configStringResult
	bools   map[string]configBoolResult
}

// configStringResult stores one string config read result.
type configStringResult struct {
	value string
	found bool
}

// configBoolResult stores one bool config read result.
type configBoolResult struct {
	value bool
	found bool
}

// Exists reports whether one fake plugin config key exists.
func (s *fakePluginConfigService) Exists(_ context.Context, key string) (bool, error) {
	if result := s.strings[key]; result.found {
		return true, nil
	}
	if result := s.bools[key]; result.found {
		return true, nil
	}
	return false, nil
}

// String returns one configured fake string value.
func (s *fakePluginConfigService) String(_ context.Context, key string, defaultValue string) (string, error) {
	result := s.strings[key]
	if !result.found {
		return defaultValue, nil
	}
	return result.value, nil
}

// Bool returns one configured fake bool value.
func (s *fakePluginConfigService) Bool(_ context.Context, key string, defaultValue bool) (bool, error) {
	result := s.bools[key]
	if !result.found {
		return defaultValue, nil
	}
	return result.value, nil
}

// fakeHostConfigHostService returns deterministic public host config values for unit
// tests.
type fakeHostConfigHostService struct {
	strings map[string]configStringResult
	bools   map[string]configBoolResult
}

// Get returns one configured fake public host config raw value.
func (s *fakeHostConfigHostService) Get(_ context.Context, key string) (*gvar.Var, error) {
	if result := s.strings[key]; result.found {
		return gvar.New(result.value), nil
	}
	if result := s.bools[key]; result.found {
		return gvar.New(result.value), nil
	}
	return nil, nil
}

// Exists reports whether one configured fake public host config key exists.
func (s *fakeHostConfigHostService) Exists(_ context.Context, key string) (bool, error) {
	if result := s.strings[key]; result.found {
		return true, nil
	}
	if result := s.bools[key]; result.found {
		return true, nil
	}
	return false, nil
}

// String returns one configured fake public host config string value.
func (s *fakeHostConfigHostService) String(_ context.Context, key string, defaultValue string) (string, error) {
	result := s.strings[key]
	if !result.found {
		return defaultValue, nil
	}
	return result.value, nil
}

// Bool returns one configured fake public host config bool value.
func (s *fakeHostConfigHostService) Bool(_ context.Context, key string, defaultValue bool) (bool, error) {
	result := s.bools[key]
	if !result.found {
		return defaultValue, nil
	}
	return result.value, nil
}

// Int returns the provided default value because tests do not configure ints.
func (s *fakeHostConfigHostService) Int(_ context.Context, _ string, defaultValue int) (int, error) {
	return defaultValue, nil
}

// Duration returns the provided default value because tests do not configure durations.
func (s *fakeHostConfigHostService) Duration(_ context.Context, _ string, defaultValue time.Duration) (time.Duration, error) {
	return defaultValue, nil
}

// fakeManifestHostService returns deterministic manifest resources for unit
// tests.
type fakeManifestHostService struct {
	texts   map[string]manifestTextResult
	profile hostCallDemoManifestProfile
}

// manifestTextResult stores one manifest text read result.
type manifestTextResult struct {
	value string
	found bool
}

// Get returns one configured fake manifest resource.
func (s *fakeManifestHostService) Get(_ context.Context, path string) ([]byte, error) {
	result := s.texts[path]
	if !result.found {
		return nil, nil
	}
	return []byte(result.value), nil
}

// Exists reports whether one configured fake manifest resource exists.
func (s *fakeManifestHostService) Exists(_ context.Context, path string) (bool, error) {
	if path == hostCallDemoManifestProfilePath {
		return true, nil
	}
	return s.texts[path].found, nil
}

// Scan copies the configured profile into the target for the expected profile
// path and key.
func (s *fakeManifestHostService) Scan(_ context.Context, path string, key string, target any) error {
	if path != hostCallDemoManifestProfilePath || strings.TrimSpace(key) != "profile" {
		return nil
	}
	profile, ok := target.(*hostCallDemoManifestProfile)
	if !ok {
		return nil
	}
	*profile = s.profile
	return nil
}

// fakeOrgHostService returns deterministic organization capability values for
// unit tests.
type fakeOrgHostService struct{}

// Status returns a deterministic organization capability status.
func (s *fakeOrgHostService) Status(_ context.Context) capmodel.CapabilityStatus {
	return capmodel.CapabilityStatus{
		CapabilityID:   orgcap.CapabilityOrgV1,
		Available:      true,
		ActiveProvider: orgcap.ProviderPluginID,
	}
}

// Available reports that the fake organization capability is active.
func (s *fakeOrgHostService) Available(_ context.Context) bool {
	return true
}

// ListUserDeptAssignments returns one deterministic current-user assignment.
func (s *fakeOrgHostService) ListUserDeptAssignments(
	_ context.Context,
	userIDs []int,
) (map[int]*orgcap.UserDeptAssignment, error) {
	result := make(map[int]*orgcap.UserDeptAssignment, len(userIDs))
	for _, userID := range userIDs {
		result[userID] = &orgcap.UserDeptAssignment{DeptID: 11, DeptName: "Engineering"}
	}
	return result, nil
}

// GetUserDeptIDs returns deterministic current-user department IDs.
func (s *fakeOrgHostService) GetUserDeptIDs(_ context.Context, _ int) ([]int, error) {
	return []int{11}, nil
}

// GetUserPostIDs returns deterministic current-user post IDs.
func (s *fakeOrgHostService) GetUserPostIDs(_ context.Context, _ int) ([]int, error) {
	return []int{21, 22}, nil
}

// fakeTenantHostService returns deterministic tenant capability values for
// unit tests.
type fakeTenantHostService struct{}

// Status returns a deterministic tenant capability status.
func (s *fakeTenantHostService) Status(_ context.Context) capmodel.CapabilityStatus {
	return capmodel.CapabilityStatus{
		CapabilityID:   tenantcap.CapabilityTenantV1,
		Available:      true,
		ActiveProvider: tenantcap.ProviderPluginID,
	}
}

// Available reports that the fake tenant capability is active.
func (s *fakeTenantHostService) Available(_ context.Context) bool {
	return true
}

// Current returns one deterministic current tenant.
func (s *fakeTenantHostService) Current(_ context.Context) tenantcap.TenantID {
	return tenantcap.TenantID(7)
}

// PlatformBypass reports that the fake request uses tenant filtering.
func (s *fakeTenantHostService) PlatformBypass(_ context.Context) bool {
	return false
}

// EnsureTenantVisible accepts the deterministic current tenant.
func (s *fakeTenantHostService) EnsureTenantVisible(_ context.Context, _ tenantcap.TenantID) error {
	return nil
}

// ListUserTenants returns deterministic current-user tenants.
func (s *fakeTenantHostService) ListUserTenants(_ context.Context, _ int) ([]tenantcap.TenantInfo, error) {
	return []tenantcap.TenantInfo{{
		ID:     tenantcap.TenantID(7),
		Code:   "tenant-demo",
		Name:   "Tenant Demo",
		Status: "active",
	}}, nil
}

// TestRunHostCallDemoConfigReadsPluginAndHostConfigValues verifies the dynamic
// demo reads plugin config and public host config values through separate
// governed clients.
func TestRunHostCallDemoConfigReadsPluginAndHostConfigValues(t *testing.T) {
	service := &serviceImpl{
		pluginConfigSvc: &fakePluginConfigService{
			strings: map[string]configStringResult{
				hostCallDemoPluginGreetingKey: {
					value: "Hello from test config",
					found: true,
				},
			},
			bools: map[string]configBoolResult{
				hostCallDemoPluginFeatureKey: {
					value: true,
					found: true,
				},
			},
		},
		hostConfigSvc: &fakeHostConfigHostService{
			strings: map[string]configStringResult{
				hostCallDemoWorkspaceKey: {
					value: "/tmp/linapro",
					found: true,
				},
				hostCallDemoI18nDefaultKey: {
					value: "zh-CN",
					found: true,
				},
			},
			bools: map[string]configBoolResult{
				hostCallDemoI18nEnabledKey: {
					value: true,
					found: true,
				},
			},
		},
	}

	payload, err := service.runHostCallDemoConfig(t.Context())
	if err != nil {
		t.Fatalf("expected config demo to succeed, got error: %v", err)
	}
	if !payload.Plugin.GreetingFound || payload.Plugin.Greeting != "Hello from test config" {
		t.Fatalf("unexpected plugin greeting payload: %#v", payload.Plugin)
	}
	if !payload.Plugin.FeatureEnabledFound || !payload.Plugin.FeatureEnabled {
		t.Fatalf("unexpected plugin feature payload: %#v", payload.Plugin)
	}
	if !payload.HostConfig.WorkspaceBasePathFound || payload.HostConfig.WorkspaceBasePath != "/tmp/linapro" {
		t.Fatalf("unexpected host workspace payload: %#v", payload.HostConfig)
	}
	if !payload.HostConfig.I18nDefaultFound || payload.HostConfig.I18nDefault != "zh-CN" {
		t.Fatalf("unexpected host i18n default payload: %#v", payload.HostConfig)
	}
	if !payload.HostConfig.I18nEnabledFound || !payload.HostConfig.I18nEnabled {
		t.Fatalf("unexpected host i18n enabled payload: %#v", payload.HostConfig)
	}
}

// TestRunHostCallDemoManifestReadsAuthorizedResources verifies the dynamic
// demo reads only the manifest resources declared for the manifest host
// service example.
func TestRunHostCallDemoManifestReadsAuthorizedResources(t *testing.T) {
	service := &serviceImpl{
		manifestSvc: &fakeManifestHostService{
			texts: map[string]manifestTextResult{
				hostCallDemoManifestConfigPath: {
					value: "demo:\n  greeting: Hello from test manifest config\n  featureEnabled: true\n",
					found: true,
				},
			},
			profile: hostCallDemoManifestProfile{
				Name:  "demo-dynamic-profile",
				Tier:  "sample",
				Owner: "linapro",
			},
		},
	}

	payload, err := service.runHostCallDemoManifest(t.Context())
	if err != nil {
		t.Fatalf("expected manifest demo to succeed, got error: %v", err)
	}
	if payload.ProfilePath != hostCallDemoManifestProfilePath || !payload.ProfileFound {
		t.Fatalf("unexpected profile path/found payload: %#v", payload)
	}
	if payload.ProfileName != "demo-dynamic-profile" ||
		payload.ProfileTier != "sample" ||
		payload.ProfileOwner != "linapro" {
		t.Fatalf("unexpected profile payload: %#v", payload)
	}
	if payload.ConfigPath != hostCallDemoManifestConfigPath || !payload.ConfigFound {
		t.Fatalf("unexpected config path/found payload: %#v", payload)
	}
	if !strings.Contains(payload.ConfigBodyPreview, "Hello from test manifest config") {
		t.Fatalf("unexpected config preview payload: %#v", payload)
	}
}

// TestRunHostCallDemoOrgTenantReadsCapabilityServices verifies the dynamic demo
// exercises organization and tenant host services through dedicated clients.
func TestRunHostCallDemoOrgTenantReadsCapabilityServices(t *testing.T) {
	service := &serviceImpl{
		orgSvc:    &fakeOrgHostService{},
		tenantSvc: &fakeTenantHostService{},
	}
	input := &HostCallDemoInput{UserID: 42}

	orgPayload, err := service.runHostCallDemoOrg(context.Background(), input)
	if err != nil {
		t.Fatalf("expected org demo to succeed, got error: %v", err)
	}
	if !orgPayload.Available || orgPayload.CapabilityID != orgcap.CapabilityOrgV1 {
		t.Fatalf("unexpected org status payload: %#v", orgPayload)
	}
	if orgPayload.AssignmentCount != 1 ||
		orgPayload.CurrentUserDeptCount != 1 ||
		orgPayload.CurrentUserPostCount != 2 {
		t.Fatalf("unexpected org projection payload: %#v", orgPayload)
	}

	tenantPayload, err := service.runHostCallDemoTenant(context.Background(), input)
	if err != nil {
		t.Fatalf("expected tenant demo to succeed, got error: %v", err)
	}
	if !tenantPayload.Available || tenantPayload.CapabilityID != tenantcap.CapabilityTenantV1 {
		t.Fatalf("unexpected tenant status payload: %#v", tenantPayload)
	}
	if tenantPayload.CurrentTenantID != 7 ||
		tenantPayload.PlatformBypass ||
		tenantPayload.UserTenantCount != 1 ||
		!tenantPayload.Visible {
		t.Fatalf("unexpected tenant projection payload: %#v", tenantPayload)
	}
}
