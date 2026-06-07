// This file verifies host-call demo helpers that do not require a running Wasm
// host service.

package dynamicservice

import (
	"context"
	"strings"
	"testing"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/tenantcap"
)

// fakeConfigHostService returns deterministic plugin config values for unit
// tests.
type fakeConfigHostService struct {
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

// String returns one configured fake string value.
func (s *fakeConfigHostService) String(key string) (string, bool, error) {
	result := s.strings[key]
	return result.value, result.found, nil
}

// Bool returns one configured fake bool value.
func (s *fakeConfigHostService) Bool(key string) (bool, bool, error) {
	result := s.bools[key]
	return result.value, result.found, nil
}

// fakeHostConfigHostService returns deterministic public host config values for unit
// tests.
type fakeHostConfigHostService struct {
	strings map[string]configStringResult
	bools   map[string]configBoolResult
}

// String returns one configured fake public host config string value.
func (s *fakeHostConfigHostService) String(key string) (string, bool, error) {
	result := s.strings[key]
	return result.value, result.found, nil
}

// Bool returns one configured fake public host config bool value.
func (s *fakeHostConfigHostService) Bool(key string) (bool, bool, error) {
	result := s.bools[key]
	return result.value, result.found, nil
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

// GetText returns one configured fake manifest text resource.
func (s *fakeManifestHostService) GetText(path string) (string, bool, error) {
	result := s.texts[path]
	return result.value, result.found, nil
}

// Scan copies the configured profile into the target for the expected profile
// path and key.
func (s *fakeManifestHostService) Scan(path string, key string, target any) (bool, error) {
	if path != hostCallDemoManifestProfilePath || strings.TrimSpace(key) != "profile" {
		return false, nil
	}
	profile, ok := target.(*hostCallDemoManifestProfile)
	if !ok {
		return false, nil
	}
	*profile = s.profile
	return true, nil
}

// fakeOrgHostService returns deterministic organization capability values for
// unit tests.
type fakeOrgHostService struct{}

// Status returns a deterministic organization capability status.
func (s *fakeOrgHostService) Status(_ context.Context) (capmodel.CapabilityStatus, error) {
	return capmodel.CapabilityStatus{
		CapabilityID:   orgcap.CapabilityOrgV1,
		Available:      true,
		ActiveProvider: orgcap.ProviderPluginID,
	}, nil
}

// Available reports that the fake organization capability is active.
func (s *fakeOrgHostService) Available(_ context.Context) (bool, error) {
	return true, nil
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
func (s *fakeTenantHostService) Status(_ context.Context) (capmodel.CapabilityStatus, error) {
	return capmodel.CapabilityStatus{
		CapabilityID:   tenantcap.CapabilityTenantV1,
		Available:      true,
		ActiveProvider: tenantcap.ProviderPluginID,
	}, nil
}

// Available reports that the fake tenant capability is active.
func (s *fakeTenantHostService) Available(_ context.Context) (bool, error) {
	return true, nil
}

// Current returns one deterministic current tenant.
func (s *fakeTenantHostService) Current(_ context.Context) (tenantcap.TenantID, error) {
	return tenantcap.TenantID(7), nil
}

// PlatformBypass reports that the fake request uses tenant filtering.
func (s *fakeTenantHostService) PlatformBypass(_ context.Context) (bool, error) {
	return false, nil
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
// demo reads plugin config and public host config values through separate host
// service clients.
func TestRunHostCallDemoConfigReadsPluginAndHostConfigValues(t *testing.T) {
	service := &serviceImpl{
		configSvc: &fakeConfigHostService{
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

	payload, err := service.runHostCallDemoConfig()
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

	payload, err := service.runHostCallDemoManifest()
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
