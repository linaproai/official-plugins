// This file verifies the dynamic AI bridge declaration helper, codecs, and
// typed owner-aware host-service client without requiring a WASI runtime.

package bridge

import (
	"strings"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-plugin-linapro-ai-core/backend/cap/aicap"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aitypes"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aiimage"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aitext"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/spi"
)

// TestHostServiceSpecDeclaresOwnerAIService verifies dynamic plugins can build
// a valid owner-aware host-service declaration from the owner SDK.
func TestHostServiceSpecDeclaresOwnerAIService(t *testing.T) {
	t.Parallel()

	spec := HostServiceSpec()
	if spec.Owner != spi.OwnerPluginID || spec.Service != spi.ServiceAI || spec.Version != spi.VersionV1 {
		t.Fatalf("unexpected owner host-service identity: %#v", spec)
	}
	if len(spec.Methods) != len(spi.MethodDescriptors()) {
		t.Fatalf("expected all descriptor methods, got %#v", spec.Methods)
	}
	if err := protocol.ValidateHostServiceSpecs([]*protocol.HostServiceSpec{spec}); err != nil {
		t.Fatalf("expected host service spec to validate: %v", err)
	}
}

// TestHostServiceSpecAcceptsSubset verifies callers can request a focused
// method subset while preserving owner identity fields.
func TestHostServiceSpecAcceptsSubset(t *testing.T) {
	t.Parallel()

	spec := HostServiceSpec(" "+spi.MethodTextGenerate+" ", "")
	if len(spec.Methods) != 1 || spec.Methods[0] != spi.MethodTextGenerate {
		t.Fatalf("expected trimmed text method subset, got %#v", spec.Methods)
	}
	if spec.Owner != spi.OwnerPluginID || spec.Service != spi.ServiceAI || spec.Version != spi.VersionV1 {
		t.Fatalf("unexpected owner host-service identity: %#v", spec)
	}
}

// TestCodecRoundTrip verifies direct DTO payloads, value requests, and host
// response envelopes round-trip through the owner bridge codec.
func TestCodecRoundTrip(t *testing.T) {
	t.Parallel()

	request := aitext.GenerateRequest{
		Purpose: "content.summary",
		Tier:    aitext.TierBasic,
		Messages: []aitext.Message{{
			Role:    aitext.MessageRoleUser,
			Content: "hello",
		}},
	}
	payload, err := MarshalRequest(request)
	if err != nil {
		t.Fatalf("marshal direct request: %v", err)
	}
	decodedRequest, err := UnmarshalRequest[aitext.GenerateRequest](payload)
	if err != nil {
		t.Fatalf("unmarshal direct request: %v", err)
	}
	if decodedRequest.Purpose != request.Purpose || decodedRequest.Messages[0].Content != "hello" {
		t.Fatalf("unexpected direct request roundtrip: %#v", decodedRequest)
	}

	statusInput := aicap.MethodStatusesInput{Methods: []aicap.MethodStatusQuery{{
		CapabilityType:   aicap.CapabilityTypeText,
		CapabilityMethod: aicap.CapabilityMethod(aitypes.CapabilityMethodTextGenerate),
	}}}
	valuePayload, err := MarshalValueRequest(statusInput)
	if err != nil {
		t.Fatalf("marshal value request: %v", err)
	}
	decodedStatusInput, err := UnmarshalValueRequest[aicap.MethodStatusesInput](valuePayload)
	if err != nil {
		t.Fatalf("unmarshal value request: %v", err)
	}
	if len(decodedStatusInput.Methods) != 1 || decodedStatusInput.Methods[0].CapabilityType != aicap.CapabilityTypeText {
		t.Fatalf("unexpected value request roundtrip: %#v", decodedStatusInput)
	}

	responsePayload, err := MarshalResponse(aitext.GenerateResponse{Text: "summary"})
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	decodedResponse, err := UnmarshalResponse[aitext.GenerateResponse](responsePayload)
	if err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if decodedResponse.Text != "summary" {
		t.Fatalf("unexpected response roundtrip: %#v", decodedResponse)
	}
}

func TestCodecRejectsMalformedAndEmptyResponses(t *testing.T) {
	t.Parallel()

	if _, err := UnmarshalRequest[aitext.GenerateRequest](nil); err == nil {
		t.Fatal("expected empty request payload to fail")
	}
	if _, err := UnmarshalValueRequest[aicap.MethodStatusesInput](nil); err == nil {
		t.Fatal("expected empty value request payload to fail")
	}
	if _, err := UnmarshalResponse[aitext.GenerateResponse]([]byte{0xff}); err == nil {
		t.Fatal("expected malformed response payload to fail")
	}
	if _, err := UnmarshalResponse[aitext.GenerateResponse](protocol.MarshalHostServiceJSONResponse(&protocol.HostServiceJSONResponse{})); err == nil {
		t.Fatal("expected empty response value to fail")
	}
}

// TestClientGenerateTextUsesOwnerHostService verifies the typed text client
// sends owner, service, version, method, and owner DTO payload correctly.
func TestClientGenerateTextUsesOwnerHostService(t *testing.T) {
	t.Parallel()

	var capturedOwner string
	var capturedService string
	var capturedVersion string
	var capturedMethod string
	client := NewClient(func(
		owner string,
		service string,
		version string,
		method string,
		_ string,
		_ string,
		payload []byte,
	) ([]byte, error) {
		capturedOwner = owner
		capturedService = service
		capturedVersion = version
		capturedMethod = method
		request, err := UnmarshalRequest[aitext.GenerateRequest](payload)
		if err != nil {
			return nil, err
		}
		if request.Purpose != "content.summary" || len(request.Messages) != 1 {
			t.Fatalf("unexpected text request payload: %#v", request)
		}
		return MarshalResponse(aitext.GenerateResponse{Text: "summary", Tier: request.Tier})
	})

	response, err := client.Text().GenerateText(t.Context(), aitext.GenerateRequest{
		Purpose: "content.summary",
		Tier:    aitext.TierBasic,
		Messages: []aitext.Message{{
			Role:    aitext.MessageRoleUser,
			Content: "hello",
		}},
	})
	if err != nil {
		t.Fatalf("generate text: %v", err)
	}
	if response.Text != "summary" || response.Tier != aitext.TierBasic {
		t.Fatalf("unexpected text response: %#v", response)
	}
	if capturedOwner != spi.OwnerPluginID ||
		capturedService != spi.ServiceAI ||
		capturedVersion != spi.VersionV1 ||
		capturedMethod != spi.MethodTextGenerate {
		t.Fatalf("unexpected owner call: owner=%s service=%s version=%s method=%s", capturedOwner, capturedService, capturedVersion, capturedMethod)
	}
}

func TestClientNilInvokerReturnsStructuredError(t *testing.T) {
	t.Parallel()

	_, err := NewClient(nil).Text().GenerateText(t.Context(), aitext.GenerateRequest{
		Purpose: "content.summary",
		Tier:    aitext.TierBasic,
		Messages: []aitext.Message{{
			Role:    aitext.MessageRoleUser,
			Content: "hello",
		}},
	})
	if err == nil || !strings.Contains(err.Error(), "host-service invoker is nil") {
		t.Fatalf("expected nil invoker error, got %v", err)
	}
}

// TestClientMethodStatusesUseJSONValueEnvelope verifies status methods use the
// generic JSON value request envelope and owner response envelope.
func TestClientMethodStatusesUseJSONValueEnvelope(t *testing.T) {
	t.Parallel()

	client := NewClient(func(
		owner string,
		service string,
		version string,
		method string,
		_ string,
		_ string,
		payload []byte,
	) ([]byte, error) {
		if owner != spi.OwnerPluginID || service != spi.ServiceAI || version != spi.VersionV1 || method != spi.MethodStatusesBatchGet {
			t.Fatalf("unexpected owner status call: owner=%s service=%s version=%s method=%s", owner, service, version, method)
		}
		request, err := UnmarshalValueRequest[aicap.MethodStatusesInput](payload)
		if err != nil {
			return nil, err
		}
		if len(request.Methods) != 1 || request.Methods[0].CapabilityType != aicap.CapabilityTypeImage {
			t.Fatalf("unexpected status request payload: %#v", request)
		}
		return MarshalResponse(aicap.MethodStatusesResult{Items: []aitypes.MethodStatus{{
			CapabilityType:   aitypes.CapabilityTypeImage,
			CapabilityMethod: aitypes.CapabilityMethodImageGenerate,
			Available:        true,
		}}})
	})

	result, err := client.MethodStatuses(t.Context(), aicap.MethodStatusesInput{Methods: []aicap.MethodStatusQuery{{
		CapabilityType:   aicap.CapabilityTypeImage,
		CapabilityMethod: aicap.CapabilityMethod(aiimage.CapabilityMethodGenerate),
	}}})
	if err != nil {
		t.Fatalf("method statuses: %v", err)
	}
	if len(result.Items) != 1 || !result.Items[0].Available || result.Items[0].CapabilityType != aitypes.CapabilityTypeImage {
		t.Fatalf("unexpected status response: %#v", result)
	}
}

// TestTextMethodStatusUsesDedicatedMethod verifies the text status shortcut
// targets text.method_status.get rather than the batch endpoint.
func TestTextMethodStatusUsesDedicatedMethod(t *testing.T) {
	t.Parallel()

	client := NewClient(func(
		_ string,
		_ string,
		_ string,
		method string,
		_ string,
		_ string,
		payload []byte,
	) ([]byte, error) {
		if method != spi.MethodTextStatusGet {
			t.Fatalf("expected text method status method, got %s", method)
		}
		request, err := UnmarshalValueRequest[aicap.MethodStatusQuery](payload)
		if err != nil {
			return nil, err
		}
		if request.CapabilityType != aicap.CapabilityTypeText ||
			request.CapabilityMethod != aicap.CapabilityMethod(aitypes.CapabilityMethodTextGenerate) {
			t.Fatalf("unexpected text status request: %#v", request)
		}
		return MarshalResponse(aitypes.MethodStatus{
			CapabilityType:   aitypes.CapabilityTypeText,
			CapabilityMethod: aitypes.CapabilityMethodTextGenerate,
			Available:        true,
		})
	})

	status := client.Text().MethodStatus(t.Context(), aitypes.CapabilityMethodTextGenerate)
	if !status.Available || status.CapabilityType != aitypes.CapabilityTypeText {
		t.Fatalf("unexpected text method status: %#v", status)
	}
}

func TestClientMethodStatusDegradesOnTransportError(t *testing.T) {
	t.Parallel()

	client := NewClient(func(
		_ string,
		_ string,
		_ string,
		_ string,
		_ string,
		_ string,
		_ []byte,
	) ([]byte, error) {
		return nil, gerror.New("upstream leaked sk-unit-secret")
	})

	status := client.Image().MethodStatus(t.Context(), aitypes.CapabilityMethodImageGenerate)
	if status.Available ||
		status.CapabilityStatus.CapabilityID != aiimage.CapabilityImageV1 ||
		status.CapabilityType != aitypes.CapabilityTypeImage ||
		status.CapabilityMethod != aitypes.CapabilityMethodImageGenerate ||
		strings.Contains(status.Reason, "sk-unit-secret") {
		t.Fatalf("expected unavailable redacted image status, got %#v", status)
	}

	textStatus := client.Text().MethodStatus(t.Context(), aitypes.CapabilityMethodTextGenerate)
	if textStatus.Available ||
		textStatus.CapabilityStatus.CapabilityID != aitext.CapabilityTextV1 ||
		strings.Contains(textStatus.Reason, "sk-unit-secret") {
		t.Fatalf("expected unavailable redacted text status, got %#v", textStatus)
	}
}
