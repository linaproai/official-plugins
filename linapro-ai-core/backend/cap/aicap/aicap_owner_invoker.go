// This file implements the runtime invoker for the AI owner descriptor. It
// builds a typed aicap.Service from the text provider factory and dispatches
// only the methods currently published by the owner descriptor.

package aicap

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/capregistry"
	"lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aitext"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aitypes"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/spi"
)

// ownerInvoker adapts a typed text provider factory to the generic plugin-owned
// capability invoker expected by the core dispatcher.
type ownerInvoker struct {
	factory aitext.ProviderFactory
}

var _ capregistry.Invoker = (*ownerInvoker)(nil)

// NewOwnerInvoker returns a runtime invoker that routes authorized dynamic host
// service calls through the public aicap.Service aggregation.
func NewOwnerInvoker(factory aitext.ProviderFactory) capregistry.Invoker {
	return &ownerInvoker{factory: factory}
}

// Invoke dispatches one authorized AI owner host-service method through the
// typed linapro-ai-core capability service.
func (i *ownerInvoker) Invoke(
	ctx context.Context,
	invocation capregistry.Invocation,
) (*capregistry.InvocationResult, error) {
	service, err := i.service(ctx, invocation)
	if err != nil {
		return ownerErrorResult(protocol.HostCallStatusInternalError, err), nil
	}

	switch strings.TrimSpace(invocation.Method) {
	case spi.MethodTextGenerate:
		request, err := decodeOwnerRequest[aitext.GenerateRequest](invocation.Payload)
		if err != nil {
			return ownerErrorResult(protocol.HostCallStatusInvalidRequest, err), nil
		}
		return ownerValueResult(service.Text().GenerateText(ctx, request))
	case spi.MethodTextStatusGet:
		request, err := decodeOwnerValueRequest[MethodStatusQuery](invocation.Payload)
		if err != nil {
			return ownerErrorResult(protocol.HostCallStatusInvalidRequest, err), nil
		}
		return ownerValueResult(
			service.Text().MethodStatus(ctx, aitypes.CapabilityMethod(request.CapabilityMethod)),
			nil,
		)
	case spi.MethodStatusesBatchGet:
		request, err := decodeOwnerValueRequest[MethodStatusesInput](invocation.Payload)
		if err != nil {
			return ownerErrorResult(protocol.HostCallStatusInvalidRequest, err), nil
		}
		return ownerValueResult(service.MethodStatuses(ctx, request))
	default:
		return ownerErrorResult(
			protocol.HostCallStatusNotFound,
			gerror.Newf("linapro-ai-core owner method not published: %s", invocation.Method),
		), nil
	}
}

func (i *ownerInvoker) service(
	ctx context.Context,
	invocation capregistry.Invocation,
) (Service, error) {
	if i == nil || i.factory == nil {
		return nil, gerror.New("linapro-ai-core text provider factory is nil")
	}
	env := aitext.ProviderEnv{PluginID: spi.OwnerPluginID}
	if invocation.Services != nil {
		env.BizCtx = invocation.Services.BizCtx()
		env.Cache = invocation.Services.Cache()
	}
	provider, err := i.factory(ctx, env)
	if err != nil {
		return nil, err
	}
	return ForPlugin(New(aitext.New(provider)), invocation.CallerPluginID), nil
}

func decodeOwnerRequest[T any](payload []byte) (T, error) {
	var out T
	if len(payload) == 0 {
		return out, gerror.New("AI owner request payload is empty")
	}
	if err := json.Unmarshal(payload, &out); err != nil {
		return out, gerror.Wrap(err, "decode AI owner request failed")
	}
	return out, nil
}

func decodeOwnerValueRequest[T any](payload []byte) (T, error) {
	var out T
	request, err := protocol.UnmarshalHostServiceJSONRequest(payload)
	if err != nil {
		return out, err
	}
	if request == nil || len(request.Value) == 0 {
		return out, gerror.New("AI owner value request is empty")
	}
	if err = json.Unmarshal(request.Value, &out); err != nil {
		return out, gerror.Wrap(err, "decode AI owner value request failed")
	}
	return out, nil
}

// ownerValueResult converts a typed owner method result into the generic
// invocation result contract and maps business/internal errors to host-call
// statuses.
func ownerValueResult[T any](value T, err error) (*capregistry.InvocationResult, error) {
	if err != nil {
		return ownerErrorResult(ownerErrorStatus(err), err), nil
	}
	content, marshalErr := json.Marshal(value)
	if marshalErr != nil {
		return ownerErrorResult(protocol.HostCallStatusInternalError, marshalErr), nil
	}
	payload := protocol.MarshalHostServiceJSONResponse(
		&protocol.HostServiceJSONResponse{Value: content},
	)
	return &capregistry.InvocationResult{
		Status:  protocol.HostCallStatusSuccess,
		Payload: payload,
	}, nil
}

// ownerErrorStatus maps structured owner errors to host-call status codes.
func ownerErrorStatus(err error) uint32 {
	if err == nil {
		return protocol.HostCallStatusInternalError
	}
	switch {
	case bizerr.Is(err, aitypes.CodeProviderUnavailable),
		bizerr.Is(err, aitext.CodeTextProviderUnavailable):
		return protocol.HostCallStatusCapabilityDenied
	case bizerr.Is(err, aitypes.CodeUnsupportedMethod):
		return protocol.HostCallStatusNotFound
	case bizerr.Is(err, capmodel.CodeCapabilityLimitExceeded):
		return protocol.HostCallStatusInvalidRequest
	default:
		if _, ok := bizerr.As(err); ok {
			return protocol.HostCallStatusInvalidRequest
		}
		return protocol.HostCallStatusInternalError
	}
}

// ownerErrorResult preserves structured bizerr metadata in host-call error
// payloads.
func ownerErrorResult(status uint32, err error) *capregistry.InvocationResult {
	response := protocol.NewHostCallErrorResponseFromError(status, err)
	return &capregistry.InvocationResult{
		Status:  response.Status,
		Payload: response.Payload,
	}
}
