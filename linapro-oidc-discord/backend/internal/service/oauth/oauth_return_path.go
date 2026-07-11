// oauth_return_path.go exposes the configured SPA login return path so the
// callback controller can compose the outbound redirect URL without carrying
// its own copy of the plugin configuration.

package oauth

// LoginReturnPath returns the SPA path the callback controller redirects to
// after the host external-login exchange completes.
func (s *serviceImpl) LoginReturnPath() string {
	if s == nil || s.configResolver == nil {
		return ""
	}
	return s.configResolver.config.LoginReturnPath
}
