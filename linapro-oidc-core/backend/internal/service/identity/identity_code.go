// This file defines linapro-oidc-core business error codes and runtime i18n metadata.

package identity

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeIdentityInvalid reports that the external identity key is invalid.
	CodeIdentityInvalid = bizerr.MustDefine(
		"PLUGIN_OIDC_CORE_IDENTITY_INVALID",
		"External identity provider and subject cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeIdentityQueryFailed reports that querying identity linkages failed.
	CodeIdentityQueryFailed = bizerr.MustDefine(
		"PLUGIN_OIDC_CORE_IDENTITY_QUERY_FAILED",
		"Failed to query external identity linkage",
		gcode.CodeInternalError,
	)
	// CodeIdentityWriteFailed reports that writing an identity linkage failed.
	CodeIdentityWriteFailed = bizerr.MustDefine(
		"PLUGIN_OIDC_CORE_IDENTITY_WRITE_FAILED",
		"Failed to write external identity linkage",
		gcode.CodeInternalError,
	)
	// CodeIdentityNotFound reports that the target linkage does not exist for
	// the current user. It deliberately does not distinguish "absent" from
	// "owned by another account" so unbinding never leaks other accounts'
	// linkage existence.
	CodeIdentityNotFound = bizerr.MustDefine(
		"PLUGIN_OIDC_CORE_IDENTITY_NOT_FOUND",
		"The external identity is not bound to the current account",
		gcode.CodeNotFound,
	)
	// CodeBindConflict reports that the external identity is already bound to
	// another account.
	CodeBindConflict = bizerr.MustDefine(
		"PLUGIN_OIDC_CORE_BIND_CONFLICT",
		"The external identity is already bound to another account",
		gcode.CodeInvalidOperation,
	)
	// CodeProvisionNotAllowed reports that auto-provisioning is not permitted
	// for this login.
	CodeProvisionNotAllowed = bizerr.MustDefine(
		"PLUGIN_OIDC_CORE_PROVISION_NOT_ALLOWED",
		"Automatic account provisioning is not enabled for this login",
		gcode.CodeInvalidOperation,
	)
	// CodeProvisionEmailConflict reports that an existing local account already
	// uses the asserted email, so silent auto-provisioning is rejected. The user
	// must sign in to the existing account and bind the identity explicitly.
	CodeProvisionEmailConflict = bizerr.MustDefine(
		"PLUGIN_OIDC_CORE_PROVISION_EMAIL_CONFLICT",
		"An account with the same email already exists. Sign in to that account and bind this identity instead",
		gcode.CodeInvalidOperation,
	)
	// CodeProvisionFailed reports that least-privilege account provisioning failed.
	CodeProvisionFailed = bizerr.MustDefine(
		"PLUGIN_OIDC_CORE_PROVISION_FAILED",
		"Failed to provision an account for the external identity",
		gcode.CodeInternalError,
	)
	// CodeUserCapabilityUnavailable reports that the host user capability is missing.
	CodeUserCapabilityUnavailable = bizerr.MustDefine(
		"PLUGIN_OIDC_CORE_USER_CAPABILITY_UNAVAILABLE",
		"Host user capability is unavailable",
		gcode.CodeInvalidOperation,
	)
)
