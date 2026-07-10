// Package aivideo defines the typed video AI capability contract exposed under
// AI().Video(). Provider operations are protocol references only, not business
// jobs or user-facing progress records.
package aivideo

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aitypes"
)

// Service defines the video AI capability consumed by host services and plugins.
type Service interface {
	// Available reports whether an active video provider is available.
	Available(ctx context.Context) bool
	// Status returns the current video AI capability activation state.
	Status(ctx context.Context) capmodel.CapabilityStatus
	// MethodStatus returns the current activation state for one video method.
	MethodStatus(ctx context.Context, method aitypes.CapabilityMethod) aitypes.MethodStatus
	// Generate executes one governed video generation request.
	Generate(ctx context.Context, request GenerateRequest) (*Response, error)
	// Edit executes one governed video editing request.
	Edit(ctx context.Context, request EditRequest) (*Response, error)
	// Extend executes one governed video extension request.
	Extend(ctx context.Context, request ExtendRequest) (*Response, error)
	// OperationGet returns one provider operation reference.
	OperationGet(ctx context.Context, request OperationGetRequest) (*Response, error)
	// OperationCancel cancels one provider operation when authorized and supported.
	OperationCancel(ctx context.Context, request OperationCancelRequest) (*aitypes.ProviderOperationRef, error)
}

const (
	// CapabilityVideoV1 identifies the versioned plugin-owned video AI capability.
	CapabilityVideoV1 = "plugin.linapro-ai-core.ai.video.v1"
	// CapabilityType identifies the video capability family.
	CapabilityType = aitypes.CapabilityTypeVideo
	// CapabilityMethodGenerate identifies video generation.
	CapabilityMethodGenerate = aitypes.CapabilityMethodVideoGenerate
	// CapabilityMethodEdit identifies video editing.
	CapabilityMethodEdit = aitypes.CapabilityMethodVideoEdit
	// CapabilityMethodExtend identifies video extension.
	CapabilityMethodExtend = aitypes.CapabilityMethodVideoExtend
	// CapabilityMethodOperationGet identifies provider operation status lookup.
	CapabilityMethodOperationGet = aitypes.CapabilityMethodVideoOperationGet
	// CapabilityMethodOperationCancel identifies provider operation cancellation.
	CapabilityMethodOperationCancel = aitypes.CapabilityMethodVideoOperationCancel
)

// GenerateRequest carries one governed video generation request.
type GenerateRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// Tier is the requested platform capability level.
	Tier aitypes.Tier `json:"tier"`
	// Prompt is the video generation instruction.
	Prompt string `json:"prompt"`
	// DurationMs optionally requests output duration in milliseconds.
	DurationMs int64 `json:"durationMs,omitempty"`
	// AspectRatio optionally requests a provider-supported aspect ratio.
	AspectRatio string `json:"aspectRatio,omitempty"`
	// Metadata carries short audit keys only.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// EditRequest carries one governed video editing request.
type EditRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// Tier is the requested platform capability level.
	Tier aitypes.Tier `json:"tier"`
	// Video references the video to edit.
	Video aitypes.AssetRef `json:"video"`
	// Prompt is the edit instruction.
	Prompt string `json:"prompt"`
	// Metadata carries short audit keys only.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ExtendRequest carries one governed video extension request.
type ExtendRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// Tier is the requested platform capability level.
	Tier aitypes.Tier `json:"tier"`
	// Video references the video to extend.
	Video aitypes.AssetRef `json:"video"`
	// Prompt is the extension instruction.
	Prompt string `json:"prompt,omitempty"`
	// DurationMs optionally requests additional duration in milliseconds.
	DurationMs int64 `json:"durationMs,omitempty"`
	// Metadata carries short audit keys only.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// OperationGetRequest carries one provider operation status lookup.
type OperationGetRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// OperationRef is the opaque provider operation reference.
	OperationRef string `json:"operationRef"`
}

// OperationCancelRequest carries one provider operation cancel request.
type OperationCancelRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// OperationRef is the opaque provider operation reference.
	OperationRef string `json:"operationRef"`
}

// Response carries video assets or a provider operation reference.
type Response struct {
	// Assets contains output video assets when available.
	Assets []aitypes.AssetResult `json:"assets,omitempty"`
	// Operation contains provider async operation state when the result is not ready.
	Operation *aitypes.ProviderOperationRef `json:"operation,omitempty"`
	// Provider contains public provider/model identity.
	Provider aitypes.ProviderInfo `json:"provider"`
	// Usage contains minimal usage details.
	Usage aitypes.Usage `json:"usage,omitempty"`
	// LatencyMs is the provider call latency in milliseconds.
	LatencyMs int `json:"latencyMs,omitempty"`
	// CreatedAt is a Unix timestamp in milliseconds.
	CreatedAt int64 `json:"createdAt,omitempty"`
}

type serviceImpl struct{}

var _ Service = (*serviceImpl)(nil)

// New creates a fallback video AI service.
func New() Service { return &serviceImpl{} }

// ForPlugin returns a plugin-scoped video service.
func ForPlugin(service Service, _ string) Service {
	if _, ok := service.(*serviceImpl); service == nil || ok {
		return &serviceImpl{}
	}
	return service
}

// Available reports whether an active video provider is available.
func (*serviceImpl) Available(context.Context) bool { return false }

// Status returns the fallback video capability status.
func (*serviceImpl) Status(context.Context) capmodel.CapabilityStatus {
	return aitypes.UnavailableStatus(CapabilityVideoV1)
}

// MethodStatus returns a fallback video method status.
func (*serviceImpl) MethodStatus(_ context.Context, method aitypes.CapabilityMethod) aitypes.MethodStatus {
	return aitypes.UnavailableMethodStatus(CapabilityVideoV1, CapabilityType, method)
}

// Generate validates the request boundary and returns a structured unavailable error.
func (*serviceImpl) Generate(_ context.Context, request GenerateRequest) (*Response, error) {
	if err := validatePurposeTier(request.Purpose, request.Tier); err != nil {
		return nil, err
	}
	return nil, unavailable(CapabilityMethodGenerate)
}

// Edit validates the request boundary and returns a structured unavailable error.
func (*serviceImpl) Edit(_ context.Context, request EditRequest) (*Response, error) {
	if err := validatePurposeTier(request.Purpose, request.Tier); err != nil {
		return nil, err
	}
	if strings.TrimSpace(request.Video.Ref) == "" {
		return nil, bizerr.NewCode(aitypes.CodeAssetRefRequired)
	}
	return nil, unavailable(CapabilityMethodEdit)
}

// Extend validates the request boundary and returns a structured unavailable error.
func (*serviceImpl) Extend(_ context.Context, request ExtendRequest) (*Response, error) {
	if err := validatePurposeTier(request.Purpose, request.Tier); err != nil {
		return nil, err
	}
	if strings.TrimSpace(request.Video.Ref) == "" {
		return nil, bizerr.NewCode(aitypes.CodeAssetRefRequired)
	}
	return nil, unavailable(CapabilityMethodExtend)
}

// OperationGet validates the request boundary and returns a structured unavailable error.
func (*serviceImpl) OperationGet(_ context.Context, request OperationGetRequest) (*Response, error) {
	if strings.TrimSpace(request.Purpose) == "" {
		return nil, bizerr.NewCode(aitypes.CodePurposeRequired)
	}
	if strings.TrimSpace(request.OperationRef) == "" {
		return nil, bizerr.NewCode(aitypes.CodeOperationRefRequired)
	}
	return nil, unavailable(CapabilityMethodOperationGet)
}

// OperationCancel validates the request boundary and returns a structured unavailable error.
func (*serviceImpl) OperationCancel(_ context.Context, request OperationCancelRequest) (*aitypes.ProviderOperationRef, error) {
	if strings.TrimSpace(request.Purpose) == "" {
		return nil, bizerr.NewCode(aitypes.CodePurposeRequired)
	}
	if strings.TrimSpace(request.OperationRef) == "" {
		return nil, bizerr.NewCode(aitypes.CodeOperationRefRequired)
	}
	return nil, unavailable(CapabilityMethodOperationCancel)
}

func validatePurposeTier(purpose string, tier aitypes.Tier) error {
	if strings.TrimSpace(purpose) == "" {
		return bizerr.NewCode(aitypes.CodePurposeRequired)
	}
	if !tier.Valid() {
		return bizerr.NewCode(aitypes.CodeTierInvalid, bizerr.P("tier", string(tier)))
	}
	return nil
}

func unavailable(method aitypes.CapabilityMethod) error {
	return bizerr.NewCode(
		aitypes.CodeProviderUnavailable,
		bizerr.P("capabilityType", string(CapabilityType)),
		bizerr.P("capabilityMethod", string(method)),
	)
}
