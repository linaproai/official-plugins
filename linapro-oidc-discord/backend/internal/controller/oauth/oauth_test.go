// This file verifies Discord OAuth callback helpers, in particular the host
// login error classification that propagates a stable runtime code into the
// frontend handoff URL fragment. The host bizerr codes are not imported
// directly because they live behind an internal package boundary; we use
// locally defined fixture codes whose only relevant attribute is the runtime
// code string, which is exactly what the classifier extracts.

package oauth

import (
	"errors"
	"testing"

	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
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
