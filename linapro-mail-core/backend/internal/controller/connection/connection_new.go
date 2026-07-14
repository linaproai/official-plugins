// Package connection implements HTTP handlers for mail connection management.
package connection

import mailsvc "lina-plugin-linapro-mail-core/backend/internal/service/mail"

// ControllerV1 binds connection endpoints.
type ControllerV1 struct {
	mailSvc mailsvc.Service
}

// NewV1 creates the connection controller.
func NewV1(mailSvc mailsvc.Service) *ControllerV1 {
	return &ControllerV1{mailSvc: mailSvc}
}
