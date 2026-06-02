// This file verifies Discord OAuth2 settings DTO validation rules. The save
// endpoint must accept an empty ClientSecret so the admin form can edit
// other fields without re-typing the secret on every save; the persistence
// layer interprets the empty value as "keep stored secret unchanged".

package v1

import (
	"context"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/util/gvalid"
)

// TestSaveSettingsAllowsEmptyClientSecret verifies that an empty ClientSecret
// does not trigger the required-field validator, preserving the
// "leave empty to keep stored secret" behavior of the admin form.
func TestSaveSettingsAllowsEmptyClientSecret(t *testing.T) {
	req := &SaveSettingsReq{
		ClientID:    "client-id",
		RedirectURI: "https://example.com/api/v1/auth/discord/callback",
	}
	err := gvalid.New().Data(req).Run(context.Background())
	if err != nil {
		t.Fatalf("expected empty client secret to pass validation, got %v", err)
	}
}

// TestSaveSettingsRequiresClientID verifies that ClientID retains the
// required validator so the configuration cannot be partially erased.
func TestSaveSettingsRequiresClientID(t *testing.T) {
	req := &SaveSettingsReq{
		RedirectURI: "https://example.com/api/v1/auth/discord/callback",
	}
	err := gvalid.New().Data(req).Run(context.Background())
	if err == nil {
		t.Fatal("expected missing client id to fail validation")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "clientid") &&
		!strings.Contains(strings.ToLower(err.Error()), "client id") &&
		!strings.Contains(strings.ToLower(err.Error()), "client_id") {
		t.Fatalf("expected validation error to mention client id, got %q", err.Error())
	}
}

// TestSaveSettingsRequiresRedirectURI verifies that RedirectURI retains the
// required validator so an empty value cannot accidentally clear the
// registered OAuth callback URL.
func TestSaveSettingsRequiresRedirectURI(t *testing.T) {
	req := &SaveSettingsReq{
		ClientID: "client-id",
	}
	err := gvalid.New().Data(req).Run(context.Background())
	if err == nil {
		t.Fatal("expected missing redirect uri to fail validation")
	}
}
