// This file verifies soft diagnostic failures localize bizerr messageKeys
// instead of exposing English-only err.Error() fallbacks.

package settings

import (
	"context"
	"errors"
	"testing"

	"lina-core/pkg/bizerr"
	"lina-plugin-linapro-mail-core/backend/cap/mailcap"
)

// stubI18n is a minimal i18ncap.Service for controller localization tests.
type stubI18n struct {
	locale string
	values map[string]string
}

func (s stubI18n) GetLocale(context.Context) string {
	if s.locale == "" {
		return "zh-CN"
	}
	return s.locale
}

func (s stubI18n) Translate(_ context.Context, key string, fallback string) string {
	if s.values != nil {
		if value, ok := s.values[key]; ok {
			return value
		}
	}
	return fallback
}

func TestLocalizeDiagnosticErrorUsesMessageKeyAndParams(t *testing.T) {
	t.Parallel()
	ctrl := NewV1(nil, stubI18n{
		values: map[string]string{
			"error.mail.transport.unavailable": "邮件传输协议 {kind} 不可用(是否未安装对应能力的插件)",
		},
	})
	err := bizerr.NewCode(mailcap.CodeMailTransportUnavailable, bizerr.P("kind", "smtp"))
	got := ctrl.localizeDiagnosticError(context.Background(), err)
	want := "邮件传输协议 smtp 不可用(是否未安装对应能力的插件)"
	if got != want {
		t.Fatalf("localized message = %q, want %q", got, want)
	}
}

func TestLocalizeDiagnosticErrorFallsBackToEnglishWithoutI18n(t *testing.T) {
	t.Parallel()
	ctrl := NewV1(nil, nil)
	err := bizerr.NewCode(mailcap.CodeMailTransportUnavailable, bizerr.P("kind", "smtp"))
	got := ctrl.localizeDiagnosticError(context.Background(), err)
	want := "Mail transport is unavailable for kind smtp (is the corresponding capability plugin not installed?)"
	if got != want {
		t.Fatalf("fallback message = %q, want %q", got, want)
	}
}

func TestLocalizeDiagnosticErrorNonBizerrUsesErrorString(t *testing.T) {
	t.Parallel()
	ctrl := NewV1(nil, stubI18n{})
	got := ctrl.localizeDiagnosticError(context.Background(), errors.New("raw failure"))
	if got != "raw failure" {
		t.Fatalf("non-bizerr message = %q", got)
	}
}

func TestLocalizeDiagnosticErrorNil(t *testing.T) {
	t.Parallel()
	ctrl := NewV1(nil, stubI18n{})
	if got := ctrl.localizeDiagnosticError(context.Background(), nil); got != "" {
		t.Fatalf("nil error message = %q", got)
	}
}
