// Package account implements HTTP handlers for mail account management.
package account

import mailsvc "lina-plugin-linapro-mail-core/backend/internal/service/mail"

// ControllerV1 binds account endpoints.
type ControllerV1 struct {
	mailSvc mailsvc.Service
}

// NewV1 creates the account controller.
func NewV1(mailSvc mailsvc.Service) *ControllerV1 {
	return &ControllerV1{mailSvc: mailSvc}
}
