// This file defines stable bizerr codes for the mail domain public contract.

package mailcap

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

// Mail domain public error codes.
var (
	// CodeMailAccountRequired reports that Account selection or default is missing.
	CodeMailAccountRequired = bizerr.MustDefine(
		"mail.account.required",
		"Mail account is required",
		gcode.CodeInvalidParameter,
	)
	// CodeMailAccountInboundNotConfigured reports Account has no inbound binding.
	CodeMailAccountInboundNotConfigured = bizerr.MustDefine(
		"mail.account.inbound_not_configured",
		"Mail account inbound connection is not configured",
		gcode.CodeInvalidOperation,
	)
	// CodeMailAccountOutboundNotConfigured reports Account has no outbound binding.
	CodeMailAccountOutboundNotConfigured = bizerr.MustDefine(
		"mail.account.outbound_not_configured",
		"Mail account outbound connection is not configured",
		gcode.CodeInvalidOperation,
	)
	// CodeMailTransportUnavailable reports no enabled SPI for the kind.
	// Hint covers common cases: the protocol plugin is not installed or not enabled.
	CodeMailTransportUnavailable = bizerr.MustDefine(
		"mail.transport.unavailable",
		"Mail transport is unavailable for kind {kind} (is the corresponding capability plugin not installed?)",
		gcode.CodeInvalidOperation,
	)
	// CodeMailTransportConflict reports multiple enabled SPI providers for one kind.
	CodeMailTransportConflict = bizerr.MustDefine(
		"mail.transport.conflict",
		"Mail transport conflict for kind {kind}: {providerIds}",
		gcode.CodeInvalidOperation,
	)
	// CodeMailConnectionNotFound reports a missing connection.
	CodeMailConnectionNotFound = bizerr.MustDefine(
		"mail.connection.not_found",
		"Mail connection not found",
		gcode.CodeNotFound,
	)
	// CodeMailInboundUnavailable reports inbound method called without SPI or binding.
	// Hint covers common cases: the IMAP/POP3 plugin is not installed or not enabled.
	CodeMailInboundUnavailable = bizerr.MustDefine(
		"mail.inbound.unavailable",
		"Mail inbound transport is unavailable (is the corresponding capability plugin not installed?)",
		gcode.CodeInvalidOperation,
	)
)
