// Package settings implements HTTP handlers for platform single-account mail settings.
package settings

import (
	"lina-core/pkg/plugin/capability/i18ncap"
	mailsvc "lina-plugin-linapro-mail-core/backend/internal/service/mail"
)

// ControllerV1 binds mail settings endpoints.
type ControllerV1 struct {
	mailSvc mailsvc.Service
	// i18nSvc localizes soft diagnostic failure messages (ok=false) for the admin UI.
	i18nSvc i18ncap.Service
}

// NewV1 creates the settings controller.
// i18nSvc may be nil in tests; missing translation then falls back to English source text.
func NewV1(mailSvc mailsvc.Service, i18nSvc i18ncap.Service) *ControllerV1 {
	return &ControllerV1{mailSvc: mailSvc, i18nSvc: i18nSvc}
}
