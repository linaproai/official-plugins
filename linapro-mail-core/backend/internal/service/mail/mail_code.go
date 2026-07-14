// This file defines mail management service error codes.

package mail

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

// Service-level error codes for connection/account management.
var (
	// CodeConnectionNameRequired reports missing connection name.
	CodeConnectionNameRequired = bizerr.MustDefine(
		"mail.connection.name_required",
		"Mail connection name is required",
		gcode.CodeInvalidParameter,
	)
	// CodeConnectionKindInvalid reports unsupported kind.
	CodeConnectionKindInvalid = bizerr.MustDefine(
		"mail.connection.kind_invalid",
		"Mail connection kind is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeConnectionHostRequired reports missing host.
	CodeConnectionHostRequired = bizerr.MustDefine(
		"mail.connection.host_required",
		"Mail connection host is required",
		gcode.CodeInvalidParameter,
	)
	// CodeConnectionPortInvalid reports invalid port.
	CodeConnectionPortInvalid = bizerr.MustDefine(
		"mail.connection.port_invalid",
		"Mail connection port is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeAccountNameRequired reports missing account name.
	CodeAccountNameRequired = bizerr.MustDefine(
		"mail.account.name_required",
		"Mail account name is required",
		gcode.CodeInvalidParameter,
	)
	// CodeAccountNotFound reports missing account.
	CodeAccountNotFound = bizerr.MustDefine(
		"mail.account.not_found",
		"Mail account not found",
		gcode.CodeNotFound,
	)
	// CodeAccountBindingInvalid reports invalid connection binding for direction/kind.
	CodeAccountBindingInvalid = bizerr.MustDefine(
		"mail.account.binding_invalid",
		"Mail account connection binding is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeIDsRequired reports empty ID batch.
	CodeIDsRequired = bizerr.MustDefine(
		"mail.ids.required",
		"IDs are required",
		gcode.CodeInvalidParameter,
	)
	// CodeSettingsUsernameRequired reports missing mailbox account/username on settings save.
	CodeSettingsUsernameRequired = bizerr.MustDefine(
		"mail.settings.username_required",
		"Mail account username is required",
		gcode.CodeInvalidParameter,
	)
	// CodeSettingsPasswordRequired reports missing password when no prior secret exists.
	CodeSettingsPasswordRequired = bizerr.MustDefine(
		"mail.settings.password_required",
		"Mail password is required for {field}",
		gcode.CodeInvalidParameter,
	)
	// CodeSettingsRecipientRequired reports missing diagnostic recipient on send-test.
	CodeSettingsRecipientRequired = bizerr.MustDefine(
		"mail.settings.recipient_required",
		"Mail test recipient is required",
		gcode.CodeInvalidParameter,
	)
	// CodeSettingsBodyRequired reports missing diagnostic body on send-test.
	CodeSettingsBodyRequired = bizerr.MustDefine(
		"mail.settings.body_required",
		"Mail test body is required",
		gcode.CodeInvalidParameter,
	)
	// CodeSettingsInboundRequired reports missing inbound protocol on receive-test.
	CodeSettingsInboundRequired = bizerr.MustDefine(
		"mail.settings.inbound_required",
		"Inbound protocol is required for receive test",
		gcode.CodeInvalidParameter,
	)
)
