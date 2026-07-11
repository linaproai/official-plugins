// Package login implements the public LDAP login HTTP handler.
package login

import (
	"context"

	v1 "lina-plugin-linapro-auth-ldap/backend/api/login/v1"
	ldapauthsvc "lina-plugin-linapro-auth-ldap/backend/internal/service/ldapauth"
)

// ControllerV1 handles public LDAP login.
type ControllerV1 struct {
	loginSvc ldapauthsvc.Service
}

// NewV1 creates the login controller.
func NewV1(loginSvc ldapauthsvc.Service) *ControllerV1 {
	return &ControllerV1{loginSvc: loginSvc}
}

// Login verifies directory credentials and returns a handoff code.
func (c *ControllerV1) Login(ctx context.Context, req *v1.LoginReq) (*v1.LoginRes, error) {
	out, err := c.loginSvc.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	return &v1.LoginRes{Handoff: out.Handoff}, nil
}
