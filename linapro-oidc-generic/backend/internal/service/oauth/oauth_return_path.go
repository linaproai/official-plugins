// oauth_return_path.go exposes the configured SPA login return path.

package oauth

// LoginReturnPath returns the SPA path used after external-login completes.
func (s *serviceImpl) LoginReturnPath() string {
	if s == nil || s.configResolver == nil {
		return ""
	}
	return s.configResolver.config.LoginReturnPath
}
