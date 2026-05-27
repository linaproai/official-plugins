// This file verifies the Google OAuth state-token signing helpers used by the
// authorization callback. The roundtrip and tamper-resistance behavior are
// covered without any network access so the tests are stable in CI.

package oauth

import (
	"strings"
	"testing"
	"time"
)

// TestStateRoundtripPreservesPayload verifies a signed state token decodes
// back into the original payload when the same client secret is supplied.
func TestStateRoundtripPreservesPayload(t *testing.T) {
	payload := StatePayload{
		StateKey:  "google",
		Nonce:     "nonce-value",
		ExpiresAt: time.Now().Add(time.Minute).Unix(),
	}
	token, err := encodeState(payload, "client-secret")
	if err != nil {
		t.Fatalf("encode oauth state failed: %v", err)
	}
	decoded, err := decodeState(token, "client-secret")
	if err != nil {
		t.Fatalf("decode oauth state failed: %v", err)
	}
	if decoded.StateKey != payload.StateKey ||
		decoded.Nonce != payload.Nonce ||
		decoded.ExpiresAt != payload.ExpiresAt {
		t.Fatalf("expected roundtripped payload %#v, got %#v", payload, decoded)
	}
}

// TestStateRejectsTamperedSignature verifies an attacker who modifies the
// MAC suffix cannot decode the state payload.
func TestStateRejectsTamperedSignature(t *testing.T) {
	token, err := encodeState(StatePayload{StateKey: "google", Nonce: "nonce"}, "client-secret")
	if err != nil {
		t.Fatalf("encode oauth state failed: %v", err)
	}
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		t.Fatalf("expected dot-separated state token, got %q", token)
	}
	tampered := parts[0] + ".deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	if _, err = decodeState(tampered, "client-secret"); err == nil {
		t.Fatal("expected tampered state token to be rejected")
	}
}

// TestStateRejectsDifferentSecret verifies state tokens cannot be replayed
// across tenants that rotated their client secret.
func TestStateRejectsDifferentSecret(t *testing.T) {
	token, err := encodeState(StatePayload{StateKey: "google", Nonce: "nonce"}, "old-secret")
	if err != nil {
		t.Fatalf("encode oauth state failed: %v", err)
	}
	if _, err = decodeState(token, "new-secret"); err == nil {
		t.Fatal("expected state token signed with old secret to be rejected by new secret")
	}
}

// TestServiceDecodeStateRejectsExpired verifies the public DecodeState helper
// rejects payloads whose deadline has already passed.
func TestServiceDecodeStateRejectsExpired(t *testing.T) {
	token, err := encodeState(StatePayload{
		StateKey:  "google",
		Nonce:     "nonce",
		ExpiresAt: time.Now().Add(-time.Minute).Unix(),
	}, "secret")
	if err != nil {
		t.Fatalf("encode oauth state failed: %v", err)
	}
	if _, err = New().DecodeState(token, "secret"); err == nil {
		t.Fatal("expected expired state token to be rejected")
	}
}

// TestBuildAuthorizeURLRequiresEnabled verifies the OAuth service refuses to
// build an authorization URL when the plugin is disabled.
func TestBuildAuthorizeURLRequiresEnabled(t *testing.T) {
	_, _, err := New().BuildAuthorizeURL(Settings{
		ClientID:     "client",
		ClientSecret: "secret",
		RedirectURI:  "https://example.com/api/v1/auth/google/callback",
		Enabled:      false,
	}, "", "google")
	if err == nil {
		t.Fatal("expected disabled provider to be rejected")
	}
}
