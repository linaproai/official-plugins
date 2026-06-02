// This file implements the framework text AI provider call path.

package ai

import (
	"context"
	"time"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/ai/aitext"
)

// GenerateText executes one framework text AI request.
func (s *serviceImpl) GenerateText(ctx context.Context, request aitext.ProviderRequest) (*aitext.GenerateResponse, error) {
	requestID := requestIDFromMetadata(request.Metadata)
	binding, err := s.resolveTierBinding(ctx, CapabilityTypeText, string(request.Tier))
	if err != nil {
		s.writeInvocation(ctx, requestID, request, nil, InvocationStatusFailed, aitext.Usage{}, 0, err)
		return nil, err
	}
	effort := binding.DefaultEffort
	if request.ThinkingEffort != nil {
		effort = string(*request.ThinkingEffort)
	}
	if !effortSupportedByBinding(binding, effort) {
		err = bizerr.NewCode(CodeThinkingEffortUnsupported)
		s.writeInvocation(ctx, requestID, request, binding, InvocationStatusFailed, aitext.Usage{}, 0, err)
		return nil, err
	}
	maxOutputTokens := request.MaxOutputTokens
	if maxOutputTokens == 0 {
		maxOutputTokens = binding.MaxOutputTokens
	}
	result, err := s.callProvider(
		ctx,
		binding,
		request.Messages,
		maxOutputTokens,
		request.Temperature,
		effort,
	)
	if err != nil {
		s.writeInvocation(ctx, requestID, request, binding, InvocationStatusFailed, aitext.Usage{}, 0, err)
		return nil, err
	}
	s.writeInvocation(
		ctx,
		requestID,
		request,
		binding,
		InvocationStatusSuccess,
		result.Usage,
		result.LatencyMs,
		nil,
	)
	generatedAt := time.Now().UnixMilli()
	actualEffort := aitext.ThinkingEffort(result.ThinkingEffort)
	return &aitext.GenerateResponse{
		Text:         result.Text,
		Tier:         request.Tier,
		ProviderName: binding.ProviderName,
		ModelName:    binding.ModelName,
		Protocol:     binding.Protocol,
		Usage:        result.Usage,
		LatencyMs:    result.LatencyMs,
		GeneratedAt:  generatedAt,
		ThinkingEffort: func() *aitext.ThinkingEffort {
			if result.ThinkingEffort == "" {
				return nil
			}
			return &actualEffort
		}(),
	}, nil
}
