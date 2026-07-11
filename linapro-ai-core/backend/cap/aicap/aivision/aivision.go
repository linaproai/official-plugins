// Package aivision defines the typed vision AI capability contract exposed
// under AI().Vision().
package aivision

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aitypes"
)

// Service defines the vision AI capability consumed by host services and plugins.
type Service interface {
	// Available reports whether an active vision provider is available.
	Available(ctx context.Context) bool
	// Status returns the current vision AI capability activation state.
	Status(ctx context.Context) capmodel.CapabilityStatus
	// MethodStatus returns the current activation state for one vision method.
	MethodStatus(ctx context.Context, method aitypes.CapabilityMethod) aitypes.MethodStatus
	// Analyze executes one governed visual analysis request.
	Analyze(ctx context.Context, request AnalyzeRequest) (*AnalyzeResponse, error)
}

const (
	// CapabilityVisionV1 identifies the versioned plugin-owned vision AI capability.
	CapabilityVisionV1 = "plugin.linapro-ai-core.ai.vision.v1"
	// CapabilityType identifies the vision capability family.
	CapabilityType = aitypes.CapabilityTypeVision
	// CapabilityMethodAnalyze identifies visual understanding.
	CapabilityMethodAnalyze = aitypes.CapabilityMethodVisionAnalyze
)

// AnalyzeRequest carries one governed visual analysis request.
type AnalyzeRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// Tier is the requested platform capability level.
	Tier aitypes.Tier `json:"tier"`
	// Images references input images, screenshots, diagrams, or frames.
	Images []aitypes.AssetRef `json:"images"`
	// Prompt is the analysis instruction.
	Prompt string `json:"prompt"`
	// Metadata carries short audit keys only.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// AnalyzeResponse carries visual analysis text and provider identity.
type AnalyzeResponse struct {
	// Text is the visual analysis result.
	Text string `json:"text"`
	// Provider contains public provider/model identity.
	Provider aitypes.ProviderInfo `json:"provider"`
	// Usage contains minimal usage details.
	Usage aitypes.Usage `json:"usage,omitempty"`
	// LatencyMs is the provider call latency in milliseconds.
	LatencyMs int `json:"latencyMs,omitempty"`
	// AnalyzedAt is a Unix timestamp in milliseconds.
	AnalyzedAt int64 `json:"analyzedAt,omitempty"`
}

type serviceImpl struct{}

var _ Service = (*serviceImpl)(nil)

// New creates a fallback vision AI service.
func New() Service { return &serviceImpl{} }

// ForPlugin returns a plugin-scoped vision service.
func ForPlugin(service Service, _ string) Service {
	if _, ok := service.(*serviceImpl); service == nil || ok {
		return &serviceImpl{}
	}
	return service
}

// Available reports whether an active vision provider is available.
func (*serviceImpl) Available(context.Context) bool { return false }

// Status returns the fallback vision capability status.
func (*serviceImpl) Status(context.Context) capmodel.CapabilityStatus {
	return aitypes.UnavailableStatus(CapabilityVisionV1)
}

// MethodStatus returns a fallback vision method status.
func (*serviceImpl) MethodStatus(_ context.Context, method aitypes.CapabilityMethod) aitypes.MethodStatus {
	return aitypes.UnavailableMethodStatus(CapabilityVisionV1, CapabilityType, method)
}

// Analyze validates the request boundary and returns a structured unavailable error.
func (*serviceImpl) Analyze(_ context.Context, request AnalyzeRequest) (*AnalyzeResponse, error) {
	if strings.TrimSpace(request.Purpose) == "" {
		return nil, bizerr.NewCode(aitypes.CodePurposeRequired)
	}
	if !request.Tier.Valid() {
		return nil, bizerr.NewCode(aitypes.CodeTierInvalid, bizerr.P("tier", string(request.Tier)))
	}
	if len(request.Images) == 0 || strings.TrimSpace(request.Images[0].Ref) == "" {
		return nil, bizerr.NewCode(aitypes.CodeAssetRefRequired)
	}
	return nil, bizerr.NewCode(
		aitypes.CodeProviderUnavailable,
		bizerr.P("capabilityType", string(CapabilityType)),
		bizerr.P("capabilityMethod", string(CapabilityMethodAnalyze)),
	)
}
