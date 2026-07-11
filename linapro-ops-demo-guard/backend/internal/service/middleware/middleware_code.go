// This file defines linapro-ops-demo-guard business error codes and runtime i18n
// metadata.

package middleware

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeDemoControlInstallManualDenied reports that demo mode install must
	// happen through startup plugin.autoEnable instead of plugin management.
	// messageKey is derived as error.demo.control.install.manual.denied.
	CodeDemoControlInstallManualDenied = bizerr.MustDefine(
		"DEMO_CONTROL_INSTALL_MANUAL_DENIED",
		"The demo control plugin can only be installed by configuring plugin.autoEnable and restarting the host; it cannot be installed from the page",
		gcode.CodeNotAuthorized,
	)
	// CodeDemoControlWriteDenied reports that demo mode rejected a write request.
	CodeDemoControlWriteDenied = bizerr.MustDefine(
		"DEMO_CONTROL_WRITE_DENIED",
		"Demo mode is enabled; write operations are disabled",
		gcode.CodeNotAuthorized,
	)
)
