// This file verifies Discord OAuth callback helpers, in particular the host
// login error classification that propagates a stable runtime code into the
// frontend handoff URL fragment. The host bizerr codes are not imported
// directly because they live behind an internal package boundary; we use
// locally defined fixture codes whose only relevant attribute is the runtime
// code string, which is exactly what the classifier extracts.

package oauth

import (
	"context"
	"errors"
	"testing"

	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
	plugincontract "lina-core/pkg/plugin/capability/contract"
	configsvc "lina-plugin-linapro-oidc-discord/backend/internal/service/config"
)

// testCodeUserNotProvisioned mirrors the host runtime code for
// "no local account linked to this external identity" without importing the
// internal host package; both definitions intentionally share the runtime
// string so this fixture stays in lockstep with the host code.
var testCodeUserNotProvisioned = bizerr.MustDefine(
	"AUTH_EXTERNAL_USER_NOT_PROVISIONED",
	"No local account is linked to this external identity",
	gcode.CodeNotAuthorized,
)

// testCodeUserDisabled mirrors the host runtime code for disabled accounts.
var testCodeUserDisabled = bizerr.MustDefine(
	"AUTH_USER_DISABLED",
	"User account is disabled",
	gcode.CodeNotAuthorized,
)

// TestClassifyHostLoginErrorExtractsRuntimeCode verifies that the bizerr
// runtime code is surfaced instead of a generic placeholder.
func TestClassifyHostLoginErrorExtractsRuntimeCode(t *testing.T) {
	if got := classifyHostLoginError(bizerr.NewCode(testCodeUserNotProvisioned)); got != "AUTH_EXTERNAL_USER_NOT_PROVISIONED" {
		t.Fatalf("classifyHostLoginError = %q, want AUTH_EXTERNAL_USER_NOT_PROVISIONED", got)
	}
	if got := classifyHostLoginError(bizerr.NewCode(testCodeUserDisabled)); got != "AUTH_USER_DISABLED" {
		t.Fatalf("classifyHostLoginError = %q, want AUTH_USER_DISABLED", got)
	}
}

// TestClassifyHostLoginErrorUnwrapsCause verifies the classifier still finds
// the structured error when callers wrap it around a lower-level cause.
func TestClassifyHostLoginErrorUnwrapsCause(t *testing.T) {
	wrapped := bizerr.WrapCode(errors.New("db unavailable"), testCodeUserNotProvisioned)
	if got := classifyHostLoginError(wrapped); got != "AUTH_EXTERNAL_USER_NOT_PROVISIONED" {
		t.Fatalf("classifyHostLoginError(wrapped) = %q, want AUTH_EXTERNAL_USER_NOT_PROVISIONED", got)
	}
}

// TestClassifyHostLoginErrorFallsBack verifies a plain non-bizerr error maps
// to a stable fallback code instead of leaking the raw error string.
func TestClassifyHostLoginErrorFallsBack(t *testing.T) {
	if got := classifyHostLoginError(errors.New("unexpected")); got != "AUTH_EXTERNAL_LOGIN_FAILED" {
		t.Fatalf("classifyHostLoginError = %q, want AUTH_EXTERNAL_LOGIN_FAILED", got)
	}
}

// TestClassifyHostLoginErrorIgnoresNil verifies a nil error produces an empty
// classification so callers can short-circuit success paths.
func TestClassifyHostLoginErrorIgnoresNil(t *testing.T) {
	if got := classifyHostLoginError(nil); got != "" {
		t.Fatalf("classifyHostLoginError(nil) = %q, want empty", got)
	}
}

// fakePluginState is an in-memory PluginStateService double whose return
// values and recorded calls let the tests assert both the value the
// controller observes and the identifier the controller queries.
type fakePluginState struct {
	providerEnabled bool
	enabled         bool
	authoritative   bool
	lastPluginID    string
}

// IsProviderEnabled records the queried pluginID and returns the configured
// platform-enabled answer. The controller must call this contract because
// it is the only one whose semantics match anonymous /auth/providers
// gating.
func (f *fakePluginState) IsProviderEnabled(_ context.Context, pluginID string) bool {
	f.lastPluginID = pluginID
	return f.providerEnabled
}

// IsEnabled records the queried pluginID and returns the configured
// business-entry visibility answer. The controller intentionally must NOT
// call this contract for OAuth flows.
func (f *fakePluginState) IsEnabled(_ context.Context, pluginID string) bool {
	f.lastPluginID = pluginID
	return f.enabled
}

// IsEnabledAuthoritative records the queried pluginID and returns the
// configured authoritative answer; included so the fake satisfies the
// PluginStateService interface even though the OAuth controller does not
// rely on it.
func (f *fakePluginState) IsEnabledAuthoritative(_ context.Context, pluginID string) bool {
	f.lastPluginID = pluginID
	return f.authoritative
}

// compile-time check that fakePluginState satisfies the host contract.
var _ plugincontract.PluginStateService = (*fakePluginState)(nil)

// TestIsProviderEnabledNilController verifies a nil controller short-circuits
// to false so even an uninitialized pointer keeps OAuth gated.
func TestIsProviderEnabledNilController(t *testing.T) {
	var c *Controller
	if c.isProviderEnabled(context.Background()) {
		t.Fatal("nil controller must report provider disabled")
	}
}

// TestIsProviderEnabledNilPluginState verifies a missing PluginState
// dependency keeps the controller fail-closed. This protects against
// construction errors that would silently leave OAuth open when the host
// could not publish the unified provider enablement contract.
func TestIsProviderEnabledNilPluginState(t *testing.T) {
	c := &Controller{}
	if c.isProviderEnabled(context.Background()) {
		t.Fatal("controller without PluginState must report provider disabled")
	}
}

// TestIsProviderEnabledTrue verifies the controller forwards the platform
// enabled snapshot's positive answer and asks for the canonical plugin id
// owned by configsvc.
func TestIsProviderEnabledTrue(t *testing.T) {
	fake := &fakePluginState{providerEnabled: true}
	c := &Controller{pluginState: fake}
	if !c.isProviderEnabled(context.Background()) {
		t.Fatal("expected provider enabled when PluginState returns true")
	}
	if fake.lastPluginID != configsvc.PluginID {
		t.Fatalf("expected PluginState to be queried with %q, got %q", configsvc.PluginID, fake.lastPluginID)
	}
}

// TestIsProviderEnabledFalse verifies the controller forwards the platform
// enabled snapshot's negative answer so disabled plugins cannot
// accidentally serve OAuth requests.
func TestIsProviderEnabledFalse(t *testing.T) {
	fake := &fakePluginState{providerEnabled: false}
	c := &Controller{pluginState: fake}
	if c.isProviderEnabled(context.Background()) {
		t.Fatal("expected provider disabled when PluginState returns false")
	}
	if fake.lastPluginID != configsvc.PluginID {
		t.Fatalf("expected PluginState to be queried with %q, got %q", configsvc.PluginID, fake.lastPluginID)
	}
}

// TestIsProviderEnabledIgnoresBusinessEntryVisibility verifies the
// controller does not fall back to IsEnabled (business-entry visibility),
// which can return false for tenants without the menu even when the
// platform snapshot has the provider enabled.
func TestIsProviderEnabledIgnoresBusinessEntryVisibility(t *testing.T) {
	fake := &fakePluginState{providerEnabled: true, enabled: false}
	c := &Controller{pluginState: fake}
	if !c.isProviderEnabled(context.Background()) {
		t.Fatal("expected provider enabled answer to come from IsProviderEnabled, not IsEnabled")
	}
}
