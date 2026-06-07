// This file implements the tier-test controller method.

package tier

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/tier/v1"

	"lina-core/pkg/apitime"
	"lina-core/pkg/plugin/capability/aicap/aitext"
	aisvc "lina-plugin-linapro-ai-core/backend/internal/service/ai"
)

// Test executes a lightweight saved or draft tier test.
func (c *ControllerV1) Test(ctx context.Context, req *v1.TestReq) (res *v1.TestRes, err error) {
	out, err := c.aiSvc.TestTier(ctx, aisvc.TierTestInput{
		CapabilityType:   req.CapabilityType,
		CapabilityMethod: req.CapabilityMethod,
		Code:             req.Code,
		ProviderId:       req.ProviderId,
		ModelId:          req.ModelId,
		ThinkingEffort:   req.ThinkingEffort,
		MaxOutputTokens:  req.MaxOutputTokens,
		Messages:         toServiceMessages(req.Messages),
	})
	if err != nil {
		return nil, err
	}
	return &v1.TestRes{
		Status:         out.Status,
		LatencyMs:      out.LatencyMs,
		ProviderName:   out.ProviderName,
		ModelName:      out.ModelName,
		Protocol:       out.Protocol,
		ThinkingEffort: out.ThinkingEffort,
		ErrorSummary:   out.ErrorSummary,
		TestedAt:       milliValue(apitime.Milli(out.TestedAt)),
	}, nil
}

// toServiceMessages converts API test messages to framework text messages.
func toServiceMessages(messages []v1.TextMessage) []aitext.Message {
	out := make([]aitext.Message, 0, len(messages))
	for _, message := range messages {
		out = append(out, aitext.Message{
			Role:    aitext.MessageRole(message.Role),
			Content: message.Content,
		})
	}
	return out
}
