// Package aiimage defines the typed image AI capability contract exposed under
// AI().Image(). It owns image-specific DTOs and fallback behavior; provider
// storage and external protocol details stay in provider plugins.
package aiimage

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aitypes"
)

// Service defines the image AI capability consumed by host services and plugins.
type Service interface {
	// Available reports whether an active image provider is available.
	Available(ctx context.Context) bool
	// Status returns the current image AI capability activation state.
	Status(ctx context.Context) capmodel.CapabilityStatus
	// MethodStatus returns the current activation state for one image method.
	MethodStatus(ctx context.Context, method aitypes.CapabilityMethod) aitypes.MethodStatus
	// Generate executes one governed image generation request.
	Generate(ctx context.Context, request GenerateRequest) (*Response, error)
	// Edit executes one governed image editing request.
	Edit(ctx context.Context, request EditRequest) (*Response, error)
}

const (
	// CapabilityImageV1 identifies the versioned plugin-owned image AI capability.
	CapabilityImageV1 = "plugin.linapro-ai-core.ai.image.v1"
)

const (
	// CapabilityType identifies the image capability family.
	CapabilityType = aitypes.CapabilityTypeImage
	// CapabilityMethodGenerate identifies image generation.
	CapabilityMethodGenerate = aitypes.CapabilityMethodImageGenerate
	// CapabilityMethodEdit identifies image editing.
	CapabilityMethodEdit = aitypes.CapabilityMethodImageEdit
)

// GenerateRequest carries one governed image generation request.
type GenerateRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// Tier is the requested platform capability level.
	Tier aitypes.Tier `json:"tier"`
	// Prompt is the text instruction used to generate images.
	Prompt string `json:"prompt"`
	// Size optionally requests a provider-supported image size.
	Size string `json:"size,omitempty"`
	// Count caps the number of generated assets.
	Count int `json:"count,omitempty"`
	// Metadata carries short audit keys and must not include prompt or response bodies.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// EditRequest carries one governed image editing request.
type EditRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// Tier is the requested platform capability level.
	Tier aitypes.Tier `json:"tier"`
	// Image references the image to edit.
	Image aitypes.AssetRef `json:"image"`
	// Mask optionally references an edit mask.
	Mask *aitypes.AssetRef `json:"mask,omitempty"`
	// Prompt is the edit instruction.
	Prompt string `json:"prompt"`
	// Count caps the number of generated assets.
	Count int `json:"count,omitempty"`
	// Metadata carries short audit keys and must not include prompt or response bodies.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Response carries image assets or a provider operation reference.
type Response struct {
	// Assets contains generated or edited image assets.
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

// serviceImpl is the fallback image service used until a provider plugin binds the method.
type serviceImpl struct{}

var _ Service = (*serviceImpl)(nil)

// New creates a fallback image AI service.
func New() Service {
	return &serviceImpl{}
}

// ForPlugin returns a plugin-scoped image service.
func ForPlugin(service Service, _ string) Service {
	if service == nil {
		return &serviceImpl{}
	}
	if _, ok := service.(*serviceImpl); ok {
		return &serviceImpl{}
	}
	return service
}

// Available reports whether an active image provider is available.
func (*serviceImpl) Available(context.Context) bool {
	return false
}

// Status returns the fallback image capability status.
func (*serviceImpl) Status(context.Context) capmodel.CapabilityStatus {
	return aitypes.UnavailableStatus(CapabilityImageV1)
}

// MethodStatus returns a fallback image method status.
func (*serviceImpl) MethodStatus(_ context.Context, method aitypes.CapabilityMethod) aitypes.MethodStatus {
	return aitypes.UnavailableMethodStatus(CapabilityImageV1, CapabilityType, method)
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
	if strings.TrimSpace(request.Image.Ref) == "" {
		return nil, bizerr.NewCode(aitypes.CodeAssetRefRequired)
	}
	return nil, unavailable(CapabilityMethodEdit)
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
