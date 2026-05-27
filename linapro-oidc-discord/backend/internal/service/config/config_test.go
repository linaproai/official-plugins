// This file verifies Discord OAuth2 configuration persistence rules against
// an in-memory PluginSettingsService double so callers do not need a live
// sys_config database to assert the secret semantics.

package config

import (
	"context"
	"errors"
	"testing"
)

// TestSaveSkipsEmptyClientSecret verifies that an empty ClientSecret leaves
// the previously stored secret intact.
func TestSaveSkipsEmptyClientSecret(t *testing.T) {
	settings := newFakePluginSettings()
	if err := settings.SetSecret(context.Background(), PluginID, keyClientSecret, "real-secret"); err != nil {
		t.Fatalf("seed secret: %v", err)
	}
	svc := New(settings)
	if err := svc.Save(context.Background(), &Settings{ClientID: "id", ClientSecret: ""}); err != nil {
		t.Fatalf("save: %v", err)
	}
	stored, _, _ := settings.read(PluginID, keyClientSecret)
	if stored != "real-secret" {
		t.Fatalf("expected stored secret to be preserved, got %q", stored)
	}
}

// TestSaveOverwritesNonEmptyClientSecret verifies that supplying a new
// secret rotates the stored value.
func TestSaveOverwritesNonEmptyClientSecret(t *testing.T) {
	settings := newFakePluginSettings()
	if err := settings.SetSecret(context.Background(), PluginID, keyClientSecret, "old"); err != nil {
		t.Fatalf("seed secret: %v", err)
	}
	svc := New(settings)
	if err := svc.Save(context.Background(), &Settings{ClientID: "id", ClientSecret: "new"}); err != nil {
		t.Fatalf("save: %v", err)
	}
	stored, _, _ := settings.read(PluginID, keyClientSecret)
	if stored != "new" {
		t.Fatalf("expected stored secret to rotate, got %q", stored)
	}
}

// TestGetReturnsDefaults verifies that absent rows resolve to the typed
// defaults so the controller never sees nil for typed primitives.
func TestGetReturnsDefaults(t *testing.T) {
	svc := New(newFakePluginSettings())
	got, err := svc.Get(context.Background())
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.EnableBackendRedirect {
		t.Fatalf("expected default EnableBackendRedirect=false")
	}
	if got.DefaultBackendRedirect != "/dashboard" {
		t.Fatalf("expected default DefaultBackendRedirect=/dashboard, got %q", got.DefaultBackendRedirect)
	}
}

// TestSaveAndGetRoundtrip verifies that typed fields survive a save→get
// roundtrip through the underlying settings store.
func TestSaveAndGetRoundtrip(t *testing.T) {
	svc := New(newFakePluginSettings())
	original := &Settings{
		ClientID:               "client-id",
		ClientSecret:           "first-secret",
		RedirectURI:            "https://example.com/api/v1/auth/discord/callback",
		EnableBackendRedirect:  true,
		DefaultBackendRedirect: "/dashboard",
		BackendRedirects:       `{"discord":"/dashboard"}`,
		Enabled:                true,
	}
	if err := svc.Save(context.Background(), original); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := svc.Get(context.Background())
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ClientID != original.ClientID {
		t.Fatalf("ClientID mismatch: got %q want %q", got.ClientID, original.ClientID)
	}
	if got.ClientSecret != original.ClientSecret {
		t.Fatalf("ClientSecret mismatch: got %q want %q", got.ClientSecret, original.ClientSecret)
	}
	if got.RedirectURI != original.RedirectURI {
		t.Fatalf("RedirectURI mismatch: got %q want %q", got.RedirectURI, original.RedirectURI)
	}
	if !got.EnableBackendRedirect {
		t.Fatal("EnableBackendRedirect lost across roundtrip")
	}
	if got.DefaultBackendRedirect != original.DefaultBackendRedirect {
		t.Fatalf("DefaultBackendRedirect mismatch: got %q", got.DefaultBackendRedirect)
	}
	if got.BackendRedirects != original.BackendRedirects {
		t.Fatalf("BackendRedirects mismatch: got %q", got.BackendRedirects)
	}
	if !got.Enabled {
		t.Fatal("Enabled lost across roundtrip")
	}
}

// TestGetMaskedClientSecret verifies that the controller-facing helper
// returns a masked projection of the stored secret without exposing the
// raw value.
func TestGetMaskedClientSecret(t *testing.T) {
	settings := newFakePluginSettings()
	if err := settings.SetSecret(context.Background(), PluginID, keyClientSecret, "secret-12345"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	svc := New(settings)
	got, err := svc.GetMaskedClientSecret(context.Background())
	if err != nil {
		t.Fatalf("masked: %v", err)
	}
	if got == "" || got == "secret-12345" {
		t.Fatalf("expected masked projection, got %q", got)
	}
}

// fakePluginSettings is an in-memory PluginSettingsService double used by
// the persistence unit tests.
type fakePluginSettings struct {
	rows map[string]string
}

// newFakePluginSettings constructs an empty fake settings store.
func newFakePluginSettings() *fakePluginSettings {
	return &fakePluginSettings{rows: map[string]string{}}
}

// GetString returns the stored value or the supplied default.
func (s *fakePluginSettings) GetString(_ context.Context, pluginID string, key string, defaultValue string) (string, error) {
	value, ok := s.rows[pluginID+"."+key]
	if !ok {
		return defaultValue, nil
	}
	return value, nil
}

// GetBool returns the parsed bool value with default fallback.
func (s *fakePluginSettings) GetBool(_ context.Context, pluginID string, key string, defaultValue bool) (bool, error) {
	value, ok := s.rows[pluginID+"."+key]
	if !ok {
		return defaultValue, nil
	}
	return value == "true", nil
}

// GetInt returns the parsed int value with default fallback.
func (s *fakePluginSettings) GetInt(_ context.Context, pluginID string, key string, defaultValue int) (int, error) {
	value, ok := s.rows[pluginID+"."+key]
	if !ok {
		return defaultValue, nil
	}
	parsed, err := parseFakeInt(value)
	if err != nil {
		return defaultValue, nil
	}
	return parsed, nil
}

// SetString stores or clears one setting.
func (s *fakePluginSettings) SetString(_ context.Context, pluginID string, key string, value string) error {
	if pluginID == "" || key == "" {
		return errors.New("fake plugin settings requires non-empty plugin id and key")
	}
	full := pluginID + "." + key
	if value == "" {
		delete(s.rows, full)
		return nil
	}
	s.rows[full] = value
	return nil
}

// SetSecret writes a secret value; empty inputs preserve the stored value.
func (s *fakePluginSettings) SetSecret(_ context.Context, pluginID string, key string, value string) error {
	if pluginID == "" || key == "" {
		return errors.New("fake plugin settings requires non-empty plugin id and key")
	}
	if value == "" {
		return nil
	}
	s.rows[pluginID+"."+key] = value
	return nil
}

// GetMaskedSecret returns the canonical masked projection of one secret.
func (s *fakePluginSettings) GetMaskedSecret(_ context.Context, pluginID string, key string) (string, error) {
	value, ok := s.rows[pluginID+"."+key]
	if !ok {
		return "", nil
	}
	if len(value) <= 6 {
		return "***", nil
	}
	return value[:3] + "***" + value[len(value)-3:], nil
}

// Delete removes one stored setting.
func (s *fakePluginSettings) Delete(_ context.Context, pluginID string, key string) error {
	delete(s.rows, pluginID+"."+key)
	return nil
}

// List returns the namespace as a flat map.
func (s *fakePluginSettings) List(_ context.Context, pluginID string) (map[string]string, error) {
	out := map[string]string{}
	prefix := pluginID + "."
	for fullKey, value := range s.rows {
		if len(fullKey) > len(prefix) && fullKey[:len(prefix)] == prefix {
			out[fullKey[len(prefix):]] = value
		}
	}
	return out, nil
}

// read is a test-only inspection helper used to assert against the fake's
// internal state.
func (s *fakePluginSettings) read(pluginID string, key string) (string, bool, error) {
	value, ok := s.rows[pluginID+"."+key]
	return value, ok, nil
}

// parseFakeInt is a minimal int parser kept inside the test file so the
// production import surface stays untouched.
func parseFakeInt(raw string) (int, error) {
	var sign int = 1
	idx := 0
	if len(raw) > 0 && (raw[0] == '+' || raw[0] == '-') {
		if raw[0] == '-' {
			sign = -1
		}
		idx = 1
	}
	if idx == len(raw) {
		return 0, errors.New("empty integer literal")
	}
	value := 0
	for ; idx < len(raw); idx++ {
		digit := raw[idx]
		if digit < '0' || digit > '9' {
			return 0, errors.New("invalid integer digit")
		}
		value = value*10 + int(digit-'0')
	}
	return sign * value, nil
}
