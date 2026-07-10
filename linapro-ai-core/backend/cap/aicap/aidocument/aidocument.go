// Package aidocument defines the typed document AI capability contract exposed
// under AI().Document().
package aidocument

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aitypes"
)

// Service defines the document AI capability consumed by host services and plugins.
type Service interface {
	// Available reports whether an active document provider is available.
	Available(ctx context.Context) bool
	// Status returns the current document AI capability activation state.
	Status(ctx context.Context) capmodel.CapabilityStatus
	// MethodStatus returns the current activation state for one document method.
	MethodStatus(ctx context.Context, method aitypes.CapabilityMethod) aitypes.MethodStatus
	// Analyze executes one governed document analysis request.
	Analyze(ctx context.Context, request AnalyzeRequest) (*Response, error)
	// Cite executes one governed citation-aware document request.
	Cite(ctx context.Context, request CiteRequest) (*Response, error)
}

const (
	// CapabilityDocumentV1 identifies the versioned plugin-owned document AI capability.
	CapabilityDocumentV1 = "plugin.linapro-ai-core.ai.document.v1"
	// CapabilityType identifies the document capability family.
	CapabilityType = aitypes.CapabilityTypeDocument
	// CapabilityMethodAnalyze identifies document analysis.
	CapabilityMethodAnalyze = aitypes.CapabilityMethodDocumentAnalyze
	// CapabilityMethodCite identifies citation-aware document analysis.
	CapabilityMethodCite = aitypes.CapabilityMethodDocumentCite
)

// AnalyzeRequest carries one governed document analysis request.
type AnalyzeRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// Tier is the requested platform capability level.
	Tier aitypes.Tier `json:"tier"`
	// Documents references input documents.
	Documents []aitypes.AssetRef `json:"documents"`
	// Prompt is the analysis instruction.
	Prompt string `json:"prompt"`
	// Metadata carries short audit keys only.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// CiteRequest carries one governed citation-aware document request.
type CiteRequest struct {
	// Purpose identifies the governed calling scenario.
	Purpose string `json:"purpose"`
	// Tier is the requested platform capability level.
	Tier aitypes.Tier `json:"tier"`
	// Documents references input documents.
	Documents []aitypes.AssetRef `json:"documents"`
	// Question is the citation-aware question.
	Question string `json:"question"`
	// Metadata carries short audit keys only.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Citation describes one cited document span without leaking document content.
type Citation struct {
	// DocumentRef references the cited document.
	DocumentRef string `json:"documentRef"`
	// Page optionally identifies the page number.
	Page int `json:"page,omitempty"`
	// Locator optionally stores a provider-neutral span locator.
	Locator string `json:"locator,omitempty"`
}

// Response carries document analysis text, citations, and provider identity.
type Response struct {
	// Text is the document analysis result.
	Text string `json:"text"`
	// Citations contains citations when available.
	Citations []Citation `json:"citations,omitempty"`
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

// New creates a fallback document AI service.
func New() Service { return &serviceImpl{} }

// ForPlugin returns a plugin-scoped document service.
func ForPlugin(service Service, _ string) Service {
	if _, ok := service.(*serviceImpl); service == nil || ok {
		return &serviceImpl{}
	}
	return service
}

// Available reports whether an active document provider is available.
func (*serviceImpl) Available(context.Context) bool { return false }

// Status returns the fallback document capability status.
func (*serviceImpl) Status(context.Context) capmodel.CapabilityStatus {
	return aitypes.UnavailableStatus(CapabilityDocumentV1)
}

// MethodStatus returns a fallback document method status.
func (*serviceImpl) MethodStatus(_ context.Context, method aitypes.CapabilityMethod) aitypes.MethodStatus {
	return aitypes.UnavailableMethodStatus(CapabilityDocumentV1, CapabilityType, method)
}

// Analyze validates the request boundary and returns a structured unavailable error.
func (*serviceImpl) Analyze(_ context.Context, request AnalyzeRequest) (*Response, error) {
	if err := validateDocumentBoundary(request.Purpose, request.Tier, request.Documents); err != nil {
		return nil, err
	}
	return nil, unavailable(CapabilityMethodAnalyze)
}

// Cite validates the request boundary and returns a structured unavailable error.
func (*serviceImpl) Cite(_ context.Context, request CiteRequest) (*Response, error) {
	if err := validateDocumentBoundary(request.Purpose, request.Tier, request.Documents); err != nil {
		return nil, err
	}
	return nil, unavailable(CapabilityMethodCite)
}

func validateDocumentBoundary(purpose string, tier aitypes.Tier, documents []aitypes.AssetRef) error {
	if strings.TrimSpace(purpose) == "" {
		return bizerr.NewCode(aitypes.CodePurposeRequired)
	}
	if !tier.Valid() {
		return bizerr.NewCode(aitypes.CodeTierInvalid, bizerr.P("tier", string(tier)))
	}
	if len(documents) == 0 || strings.TrimSpace(documents[0].Ref) == "" {
		return bizerr.NewCode(aitypes.CodeAssetRefRequired)
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
