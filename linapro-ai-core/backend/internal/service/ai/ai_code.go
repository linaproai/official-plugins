// This file defines linapro-ai-core business error codes.

package ai

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodePlatformRequired reports that a management API was called outside platform context.
	CodePlatformRequired = bizerr.MustDefine(
		"AI_CORE_PLATFORM_REQUIRED",
		"Smart Center requires platform context",
		gcode.CodeNotAuthorized,
	)
	// CodeProviderNotFound reports that the requested provider is absent.
	CodeProviderNotFound = bizerr.MustDefine(
		"AI_CORE_PROVIDER_NOT_FOUND",
		"AI provider does not exist",
		gcode.CodeNotFound,
	)
	// CodeProviderInUse reports that a provider is referenced by a tier binding.
	CodeProviderInUse = bizerr.MustDefine(
		"AI_CORE_PROVIDER_IN_USE",
		"AI provider is used by a capability tier",
		gcode.CodeInvalidOperation,
	)
	// CodeProviderEndpointNotFound reports that the requested provider endpoint is absent.
	CodeProviderEndpointNotFound = bizerr.MustDefine(
		"AI_CORE_PROVIDER_ENDPOINT_NOT_FOUND",
		"AI provider endpoint does not exist",
		gcode.CodeNotFound,
	)
	// CodeProviderEndpointInUse reports that a provider endpoint is referenced by a model or capability.
	CodeProviderEndpointInUse = bizerr.MustDefine(
		"AI_CORE_PROVIDER_ENDPOINT_IN_USE",
		"AI provider endpoint is used by a model capability",
		gcode.CodeInvalidOperation,
	)
	// CodeProviderProtocolRequired reports missing provider protocol endpoint configuration.
	CodeProviderProtocolRequired = bizerr.MustDefine(
		"AI_CORE_PROVIDER_PROTOCOL_REQUIRED",
		"Provider requires an enabled endpoint for the selected protocol",
		gcode.CodeInvalidParameter,
	)
	// CodeProviderHTTPError reports a provider-side HTTP failure without exposing response bodies.
	CodeProviderHTTPError = bizerr.MustDefine(
		"AI_CORE_PROVIDER_HTTP_ERROR",
		"AI provider returned HTTP {status}",
		gcode.CodeInvalidOperation,
	)
	// CodeModelNotFound reports that the requested model is absent.
	CodeModelNotFound = bizerr.MustDefine(
		"AI_CORE_MODEL_NOT_FOUND",
		"AI model does not exist",
		gcode.CodeNotFound,
	)
	// CodeModelInUse reports that a model is referenced by a tier binding.
	CodeModelInUse = bizerr.MustDefine(
		"AI_CORE_MODEL_IN_USE",
		"AI model is used by a capability tier",
		gcode.CodeInvalidOperation,
	)
	// CodeTierNotFound reports that the requested tier is absent.
	CodeTierNotFound = bizerr.MustDefine(
		"AI_CORE_TIER_NOT_FOUND",
		"AI tier does not exist",
		gcode.CodeNotFound,
	)
	// CodeTierBindingUnavailable reports that a tier has no usable binding.
	CodeTierBindingUnavailable = bizerr.MustDefine(
		"AI_CORE_TIER_BINDING_UNAVAILABLE",
		"AI tier is not configured with an enabled provider and model",
		gcode.CodeInvalidOperation,
	)
	// CodeThinkingEffortUnsupported reports a model capability mismatch.
	CodeThinkingEffortUnsupported = bizerr.MustDefine(
		"AI_CORE_THINKING_EFFORT_UNSUPPORTED",
		"The selected model does not support this thinking effort",
		gcode.CodeInvalidParameter,
	)
	// CodeInvocationNotFound reports that the requested invocation log is absent.
	CodeInvocationNotFound = bizerr.MustDefine(
		"AI_CORE_INVOCATION_NOT_FOUND",
		"AI invocation log does not exist",
		gcode.CodeNotFound,
	)
	// CodeProviderOperationNotFound reports that the requested provider operation is absent.
	CodeProviderOperationNotFound = bizerr.MustDefine(
		"AI_CORE_PROVIDER_OPERATION_NOT_FOUND",
		"AI provider operation does not exist",
		gcode.CodeNotFound,
	)
	// CodeRequestInvalid reports invalid Smart Center configuration input.
	CodeRequestInvalid = bizerr.MustDefine(
		"AI_CORE_REQUEST_INVALID",
		"AI configuration request is invalid",
		gcode.CodeInvalidParameter,
	)
)
