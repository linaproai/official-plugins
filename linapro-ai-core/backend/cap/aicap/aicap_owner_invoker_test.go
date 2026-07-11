// This file verifies the AI owner invoker routes published methods through
// aicap.Service and rejects unpublished multimodal methods.
package aicap

import (
	"context"
	"encoding/json"
	"testing"

	"lina-core/pkg/plugin/capability/capregistry"
	"lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aitext"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aitypes"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/spi"
)

func TestProviderDescriptorRejectsNilFactory(t *testing.T) {
	t.Parallel()

	if _, err := ProviderDescriptor(nil); err == nil {
		t.Fatal("expected nil text provider factory to fail")
	}
}

func TestProviderDescriptorPublishesInvokerForRegisteredMethods(t *testing.T) {
	t.Parallel()

	descriptor, err := ProviderDescriptor(func(context.Context, aitext.ProviderEnv) (aitext.Provider, error) {
		return nil, nil
	})
	if err != nil {
		t.Fatalf("expected descriptor, got %v", err)
	}
	if descriptor.Invoker == nil {
		t.Fatal("provider descriptor must attach the runtime AI owner invoker")
	}

	registry := capregistry.NewRegistry()
	if err = registry.Register(descriptor); err != nil {
		t.Fatalf("register descriptor: %v", err)
	}
	for _, method := range []string{
		spi.MethodTextGenerate,
		spi.MethodTextStatusGet,
		spi.MethodStatusesBatchGet,
	} {
		registered, ok := registry.LookupMethod(spi.OwnerPluginID, spi.ServiceAI, spi.VersionV1, method)
		if !ok {
			t.Fatalf("expected method %s to be indexed", method)
		}
		if registered.Invoker == nil {
			t.Fatalf("expected method %s to retain AI owner invoker", method)
		}
	}
	if _, ok := registry.LookupMethod(spi.OwnerPluginID, spi.ServiceAI, spi.VersionV1, spi.MethodImageGenerate); ok {
		t.Fatal("unpublished multimodal methods must not enter the owner registry")
	}
}

func TestOwnerInvokerDispatchesTypedAIService(t *testing.T) {
	t.Parallel()

	provider := &fakeTextProvider{}
	descriptor, err := ProviderDescriptor(func(_ context.Context, env aitext.ProviderEnv) (aitext.Provider, error) {
		if env.PluginID != spi.OwnerPluginID {
			t.Fatalf("unexpected provider env plugin id: %s", env.PluginID)
		}
		return provider, nil
	})
	if err != nil {
		t.Fatalf("create descriptor: %v", err)
	}

	generatePayload, err := json.Marshal(aitext.GenerateRequest{
		Purpose: "content.summary",
		Tier:    aitext.TierBasic,
		Messages: []aitext.Message{
			{Role: aitext.MessageRoleUser, Content: "hello"},
		},
	})
	if err != nil {
		t.Fatalf("marshal generate request: %v", err)
	}
	generateResult, err := descriptor.Invoker.Invoke(context.Background(), capregistry.Invocation{
		CallerPluginID: "dynamic-consumer",
		OwnerPluginID:  spi.OwnerPluginID,
		Service:        spi.ServiceAI,
		Version:        spi.VersionV1,
		Method:         spi.MethodTextGenerate,
		Payload:        generatePayload,
	})
	if err != nil {
		t.Fatalf("invoke text generate: %v", err)
	}
	if generateResult == nil || generateResult.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("unexpected generate result: %#v", generateResult)
	}
	var generateResponse aitext.GenerateResponse
	decodeAIInvokerResponse(t, generateResult.Payload, &generateResponse)
	if generateResponse.Text != "generated" || provider.sourcePluginID != "dynamic-consumer" {
		t.Fatalf("unexpected generate response=%#v source=%q", generateResponse, provider.sourcePluginID)
	}

	statusPayload, err := json.Marshal(MethodStatusQuery{
		CapabilityType:   CapabilityTypeText,
		CapabilityMethod: CapabilityMethod(aitypes.CapabilityMethodTextGenerate),
	})
	if err != nil {
		t.Fatalf("marshal status request: %v", err)
	}
	statusResult, err := descriptor.Invoker.Invoke(context.Background(), capregistry.Invocation{
		CallerPluginID: "dynamic-consumer",
		OwnerPluginID:  spi.OwnerPluginID,
		Service:        spi.ServiceAI,
		Version:        spi.VersionV1,
		Method:         spi.MethodTextStatusGet,
		Payload: protocol.MarshalHostServiceJSONRequest(&protocol.HostServiceJSONRequest{
			Value: statusPayload,
		}),
	})
	if err != nil {
		t.Fatalf("invoke text method status: %v", err)
	}
	if statusResult == nil || statusResult.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("unexpected status result: %#v", statusResult)
	}
	var status aitypes.MethodStatus
	decodeAIInvokerResponse(t, statusResult.Payload, &status)
	if !status.Available || status.CapabilityType != aitypes.CapabilityTypeText {
		t.Fatalf("unexpected method status: %#v", status)
	}

	unpublished, err := descriptor.Invoker.Invoke(context.Background(), capregistry.Invocation{
		CallerPluginID: "dynamic-consumer",
		OwnerPluginID:  spi.OwnerPluginID,
		Service:        spi.ServiceAI,
		Version:        spi.VersionV1,
		Method:         spi.MethodImageGenerate,
		Payload:        []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("invoke unpublished method: %v", err)
	}
	if unpublished == nil || unpublished.Status != protocol.HostCallStatusNotFound {
		t.Fatalf("expected unpublished method not found, got %#v", unpublished)
	}
}

type fakeTextProvider struct {
	sourcePluginID string
}

func (p *fakeTextProvider) GenerateText(_ context.Context, request aitext.ProviderRequest) (*aitext.GenerateResponse, error) {
	p.sourcePluginID = request.SourcePluginID
	return &aitext.GenerateResponse{
		Text:         "generated",
		Tier:         request.Tier,
		ProviderName: "fake",
		ModelName:    "fake-model",
		Protocol:     "test",
	}, nil
}

func decodeAIInvokerResponse(t *testing.T, payload []byte, out any) {
	t.Helper()
	response, err := protocol.UnmarshalHostServiceJSONResponse(payload)
	if err != nil {
		t.Fatalf("decode host-service response: %v", err)
	}
	if err = json.Unmarshal(response.Value, out); err != nil {
		t.Fatalf("decode AI response JSON: %v", err)
	}
}
