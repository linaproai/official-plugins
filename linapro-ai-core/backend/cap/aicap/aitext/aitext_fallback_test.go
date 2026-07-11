// This file verifies text AI capability fallback, validation, and owner
// provider delegation without depending on plugin host runtime managers.

package aitext

import (
	"context"
	"testing"
	"time"

	"lina-core/pkg/bizerr"
)

func TestGenerateTextReturnsUnavailableWithoutProvider(t *testing.T) {
	t.Parallel()

	service := NewUnavailable()
	_, err := service.GenerateText(context.Background(), validGenerateRequest())
	if !bizerr.Is(err, CodeTextProviderUnavailable) {
		t.Fatalf("expected provider unavailable error, got %v", err)
	}
	if service.Available(context.Background()) {
		t.Fatal("expected text AI capability to be unavailable without provider")
	}
}

func TestGenerateTextValidatesTierAndThinkingEffort(t *testing.T) {
	t.Parallel()

	service := NewUnavailable()
	request := validGenerateRequest()
	request.Tier = Tier("custom")
	_, err := service.GenerateText(context.Background(), request)
	if !bizerr.Is(err, CodeTextTierInvalid) {
		t.Fatalf("expected invalid tier error, got %v", err)
	}

	request = validGenerateRequest()
	invalidEffort := ThinkingEffort("extreme")
	request.ThinkingEffort = &invalidEffort
	_, err = service.GenerateText(context.Background(), request)
	if !bizerr.Is(err, CodeTextThinkingEffortInvalid) {
		t.Fatalf("expected invalid thinking effort error, got %v", err)
	}
}

func TestGenerateRequestCapabilityIdentity(t *testing.T) {
	t.Parallel()

	request := validGenerateRequest()
	if request.CapabilityType() != CapabilityTypeText {
		t.Fatalf("expected capability type %q, got %q", CapabilityTypeText, request.CapabilityType())
	}
	if request.CapabilityMethod() != CapabilityMethodGenerate {
		t.Fatalf("expected capability method %q, got %q", CapabilityMethodGenerate, request.CapabilityMethod())
	}
}

func TestGenerateTextDelegatesToActiveProvider(t *testing.T) {
	t.Parallel()

	service := New(fakeProvider{})
	response, err := service.GenerateText(context.Background(), validGenerateRequest())
	if err != nil {
		t.Fatalf("expected active provider success, got %v", err)
	}
	if response == nil || response.Text != "generated text" || response.Tier != TierStandard {
		t.Fatalf("unexpected response: %#v", response)
	}
	status := service.Status(context.Background())
	if !status.Available || status.ActiveProvider != ProviderPluginID {
		t.Fatalf("expected active provider status, got %#v", status)
	}
}

func TestForPluginInjectsSourcePluginID(t *testing.T) {
	t.Parallel()

	var seenSourcePluginID string
	service := ForPlugin(New(fakeProviderFunc(func(_ context.Context, request ProviderRequest) (*GenerateResponse, error) {
		seenSourcePluginID = request.SourcePluginID
		return fakeProvider{}.GenerateText(context.Background(), request)
	})), " source-plugin ")
	if _, err := service.GenerateText(context.Background(), validGenerateRequest()); err != nil {
		t.Fatalf("expected active provider success, got %v", err)
	}
	if seenSourcePluginID != "source-plugin" {
		t.Fatalf("expected scoped source plugin id, got %q", seenSourcePluginID)
	}
}

func validGenerateRequest() GenerateRequest {
	return GenerateRequest{
		Purpose: "test.summary",
		Tier:    TierStandard,
		Messages: []Message{
			{Role: MessageRoleUser, Content: "hello"},
		},
		MaxOutputTokens: 128,
	}
}

type fakeProvider struct{}

func (fakeProvider) GenerateText(context.Context, ProviderRequest) (*GenerateResponse, error) {
	return &GenerateResponse{
		Text:         "generated text",
		Tier:         TierStandard,
		ProviderName: "Fake",
		ModelName:    "fake-model",
		Protocol:     "test",
		Usage:        Usage{InputTokens: 1, OutputTokens: 2},
		GeneratedAt:  time.Now().UnixMilli(),
	}, nil
}

type fakeProviderFunc func(context.Context, ProviderRequest) (*GenerateResponse, error)

func (f fakeProviderFunc) GenerateText(ctx context.Context, request ProviderRequest) (*GenerateResponse, error) {
	return f(ctx, request)
}
