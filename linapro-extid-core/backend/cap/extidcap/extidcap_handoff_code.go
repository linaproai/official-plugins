// extidcap_handoff_code.go defines bizerr codes for SPA login handoff exchange
// failures. Codes are part of the public contract so protocol plugins and HTTP
// adapters can match typed errors without importing owner internal packages.

package extidcap

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

// CodeLoginHandoffInvalid reports a missing, expired, already-consumed, or
// unbound handoff exchange attempt.
var CodeLoginHandoffInvalid = bizerr.MustDefine(
	"PLUGIN_EXTID_CORE_LOGIN_HANDOFF_INVALID",
	"External login handoff is invalid or expired",
	gcode.CodeNotAuthorized,
)

func errHandoffInvalid() error {
	return bizerr.NewCode(CodeLoginHandoffInvalid)
}
