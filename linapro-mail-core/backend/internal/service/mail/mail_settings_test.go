// This file unit-tests platform settings helpers without a database.

package mail

import (
	"testing"

	"lina-plugin-linapro-mail-core/backend/internal/model/entity"
)

func TestNormalizeInboundKind(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"":     InboundKindNone,
		"none": InboundKindNone,
		"IMAP": "imap",
		"pop3": "pop3",
		"smtp": "smtp",
	}
	for in, want := range cases {
		if got := normalizeInboundKind(in); got != want {
			t.Fatalf("normalizeInboundKind(%q)=%q want %q", in, got, want)
		}
	}
}

func TestResolveSecret(t *testing.T) {
	t.Parallel()
	if got := resolveSecret("  new  ", nil); got != "new" {
		t.Fatalf("submitted secret: got %q", got)
	}
	prev := &entity.Connection{SecretRef: "stored"}
	if got := resolveSecret("", prev); got != "stored" {
		t.Fatalf("empty submit keeps stored: got %q", got)
	}
	if got := resolveSecret("override", prev); got != "override" {
		t.Fatalf("override: got %q", got)
	}
}

func TestNormalizePlatformSettingsInput(t *testing.T) {
	t.Parallel()
	got := normalizePlatformSettingsInput(PlatformSettingsInput{
		SmtpUsername: "  user@example.com  ",
		SmtpHost:     "  smtp.example.com ",
	})
	if got.SmtpUsername != "user@example.com" {
		t.Fatalf("username trim: %q", got.SmtpUsername)
	}
	if got.FromAddress != "user@example.com" {
		t.Fatalf("from defaults to account: %q", got.FromAddress)
	}
	if got.InboundUsername != "user@example.com" {
		t.Fatalf("inbound username defaults to account: %q", got.InboundUsername)
	}
	got = normalizePlatformSettingsInput(PlatformSettingsInput{
		SmtpUsername: "user@example.com",
		FromAddress:  "noreply@example.com",
	})
	if got.FromAddress != "noreply@example.com" {
		t.Fatalf("explicit from kept: %q", got.FromAddress)
	}
}

func TestResolvePlatformAccountName(t *testing.T) {
	t.Parallel()
	if got := resolvePlatformAccountName(PlatformSettingsInput{Name: "custom"}); got != "custom" {
		t.Fatalf("explicit name: %q", got)
	}
	if got := resolvePlatformAccountName(PlatformSettingsInput{
		FromAddress:  "from@example.com",
		SmtpUsername: "user@example.com",
	}); got != "from@example.com" {
		t.Fatalf("derived from address: %q", got)
	}
	if got := resolvePlatformAccountName(PlatformSettingsInput{
		SmtpUsername: "user@example.com",
	}); got != "user@example.com" {
		t.Fatalf("derived from username: %q", got)
	}
}

func TestValidatePlatformSettingsSave(t *testing.T) {
	t.Parallel()
	if err := validatePlatformSettingsSave(normalizePlatformSettingsInput(PlatformSettingsInput{})); err == nil {
		t.Fatal("expected username required")
	}
	if err := validatePlatformSettingsSave(normalizePlatformSettingsInput(PlatformSettingsInput{
		SmtpUsername: "user@example.com",
	})); err == nil {
		t.Fatal("expected smtp host required")
	}
	// Empty fromAddress is valid after normalize (defaults to account).
	if err := validatePlatformSettingsSave(normalizePlatformSettingsInput(PlatformSettingsInput{
		SmtpUsername: "user@example.com",
		SmtpHost:     "smtp.example.com",
		SmtpPort:     587,
		InboundKind:  "none",
	})); err != nil {
		t.Fatalf("valid input with default from failed: %v", err)
	}
	if err := validatePlatformSettingsSave(normalizePlatformSettingsInput(PlatformSettingsInput{
		SmtpUsername: "user@example.com",
		FromAddress:  "noreply@example.com",
		SmtpHost:     "smtp.example.com",
		SmtpPort:     587,
		InboundKind:  "none",
	})); err != nil {
		t.Fatalf("valid input with explicit from failed: %v", err)
	}
}

func TestFirstNonEmpty(t *testing.T) {
	t.Parallel()
	if got := firstNonEmpty("", "  ", "user"); got != "user" {
		t.Fatalf("got %q", got)
	}
}

func TestSendPlatformTestMailValidation(t *testing.T) {
	t.Parallel()
	svc := &serviceImpl{}
	// Missing recipient must fail closed before dialing SMTP.
	_, err := svc.SendPlatformTestMail(t.Context(), PlatformSettingsInput{
		SmtpUsername: "user@example.com",
		SmtpHost:     "smtp.example.com",
		SmtpPort:     587,
		SmtpPassword: "secret",
	}, "", "subject", "body")
	if err == nil {
		t.Fatal("expected recipient required")
	}
	_, err = svc.SendPlatformTestMail(t.Context(), PlatformSettingsInput{
		SmtpUsername: "user@example.com",
		SmtpHost:     "smtp.example.com",
		SmtpPort:     587,
		SmtpPassword: "secret",
	}, "ops@example.com", "subject", "")
	if err == nil {
		t.Fatal("expected body required")
	}
	_, err = svc.SendPlatformTestMail(t.Context(), PlatformSettingsInput{
		SmtpHost:     "smtp.example.com",
		SmtpPort:     587,
		SmtpPassword: "secret",
	}, "ops@example.com", "subject", "hello")
	if err == nil {
		t.Fatal("expected username required")
	}
}

func TestPlatformReceiveValidation(t *testing.T) {
	t.Parallel()
	svc := &serviceImpl{}
	// inbound none must fail closed before dialing or DB access.
	_, err := svc.TestPlatformReceive(t.Context(), PlatformSettingsInput{
		SmtpUsername: "user@example.com",
		SmtpPassword: "secret",
		InboundKind:  InboundKindNone,
		InboundHost:  "imap.example.com",
		InboundPort:  993,
	})
	if err == nil {
		t.Fatal("expected inbound required")
	}
	// Missing host must fail closed before DB access.
	_, err = svc.TestPlatformReceive(t.Context(), PlatformSettingsInput{
		SmtpUsername: "user@example.com",
		SmtpPassword: "secret",
		InboundKind:  "imap",
		InboundPort:  993,
	})
	if err == nil {
		t.Fatal("expected host required")
	}
	// Missing username must fail closed before DB access.
	_, err = svc.TestPlatformReceive(t.Context(), PlatformSettingsInput{
		SmtpPassword: "secret",
		InboundKind:  "imap",
		InboundHost:  "imap.example.com",
		InboundPort:  993,
	})
	if err == nil {
		t.Fatal("expected username required")
	}
	// Invalid port must fail closed before DB access.
	_, err = svc.TestPlatformReceive(t.Context(), PlatformSettingsInput{
		SmtpUsername: "user@example.com",
		SmtpPassword: "secret",
		InboundKind:  "imap",
		InboundHost:  "imap.example.com",
		InboundPort:  0,
	})
	if err == nil {
		t.Fatal("expected port invalid")
	}
}
