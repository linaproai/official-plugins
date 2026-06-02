// This file contains shared validation, projection, and masking helpers for
// the Smart Center service.

package ai

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/ai/aitext"
	"lina-plugin-linapro-ai-core/backend/internal/model/entity"
)

// ensurePlatform verifies that management APIs are only used from platform context.
func (s *serviceImpl) ensurePlatform(ctx context.Context) error {
	if s == nil || s.bizCtxSvc == nil {
		return nil
	}
	current := s.bizCtxSvc.Current(ctx)
	if current.TenantID == 0 && current.PlatformBypass && !current.ActingAsTenant {
		return nil
	}
	return bizerr.NewCode(CodePlatformRequired)
}

// normalizePage applies stable pagination defaults and bounds.
func normalizePage(pageNum int, pageSize int) (int, int) {
	if pageNum <= 0 {
		pageNum = defaultPageNum
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return pageNum, pageSize
}

// normalizeEnabled coerces non-1 values to disabled for persisted flags.
func normalizeEnabled(value int) int {
	if value == enabledYes {
		return enabledYes
	}
	return enabledNo
}

// normalizeCapabilityType returns the default supported capability type.
func normalizeCapabilityType(value string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if trimmed == "" {
		return CapabilityTypeText
	}
	return trimmed
}

// normalizeProtocol returns a supported provider protocol or an empty string.
func normalizeProtocol(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case ProtocolOpenAI:
		return ProtocolOpenAI
	case ProtocolAnthropic:
		return ProtocolAnthropic
	default:
		return ""
	}
}

// normalizeTierCode returns one fixed tier code or an empty string.
func normalizeTierCode(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case TierCodeBasic:
		return TierCodeBasic
	case TierCodeStandard:
		return TierCodeStandard
	case TierCodeAdvanced:
		return TierCodeAdvanced
	default:
		return ""
	}
}

// normalizeEffort validates one optional thinking effort string.
func normalizeEffort(value string) (string, error) {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if trimmed == "" {
		return "", nil
	}
	effort := aitext.ThinkingEffort(trimmed)
	if !effort.Valid() {
		return "", bizerr.NewCode(CodeThinkingEffortUnsupported)
	}
	return trimmed, nil
}

// normalizeEfforts validates, de-duplicates, and serializes thinking efforts.
func normalizeEfforts(values []string) (string, []string, error) {
	if len(values) == 0 {
		return "", nil, nil
	}
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		effort, err := normalizeEffort(value)
		if err != nil {
			return "", nil, err
		}
		if effort == "" {
			continue
		}
		if _, ok := seen[effort]; ok {
			continue
		}
		seen[effort] = struct{}{}
		out = append(out, effort)
	}
	return strings.Join(out, ","), out, nil
}

// splitEfforts converts persisted comma-separated efforts into API values.
func splitEfforts(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// effortSupported reports whether effort is supported by the model declaration.
func effortSupported(model *entity.Model, effort string) bool {
	normalized, err := normalizeEffort(effort)
	if err != nil || normalized == "" {
		return err == nil
	}
	if model == nil || model.SupportsThinking != enabledYes {
		return false
	}
	for _, supported := range splitEfforts(model.SupportedEfforts) {
		if supported == normalized {
			return true
		}
	}
	return false
}

// maskSecretRef returns a stable masked projection of a secret reference.
func maskSecretRef(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if strings.Contains(trimmed, "***") {
		return trimmed
	}
	sum := sha256.Sum256([]byte(trimmed))
	suffix := hex.EncodeToString(sum[:])[:8]
	return "secret-***" + suffix
}

// shouldKeepExistingSecret reports whether an update input should retain the stored secret reference.
func shouldKeepExistingSecret(value string) bool {
	trimmed := strings.TrimSpace(value)
	return trimmed == "" || strings.Contains(trimmed, "***")
}

// sanitizeErrorSummary masks common authentication markers and bounds error text.
func sanitizeErrorSummary(err error) string {
	if err == nil {
		return ""
	}
	return sanitizeText(err.Error(), 512)
}

// sanitizeText masks sensitive text and truncates it to maxLen runes.
func sanitizeText(value string, maxLen int) string {
	text := strings.TrimSpace(value)
	replacements := []string{"authorization", "api-key", "apikey", "bearer", "sk-"}
	lower := strings.ToLower(text)
	for _, marker := range replacements {
		if strings.Contains(lower, marker) {
			text = "[redacted sensitive provider error]"
			break
		}
	}
	if maxLen <= 0 || len(text) <= maxLen {
		return text
	}
	return text[:maxLen]
}

// requestIDFromMetadata returns a bounded request ID or a generated fallback.
func requestIDFromMetadata(metadata map[string]string) string {
	if value := strings.TrimSpace(metadata["requestId"]); value != "" {
		return sanitizeText(value, 64)
	}
	return "ai-" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

// providerToItem converts one provider entity into a service projection.
func providerToItem(row *entity.Provider, modelCount int, enabledModelCount int) *ProviderItem {
	if row == nil {
		return nil
	}
	return &ProviderItem{
		Id:                row.Id,
		Name:              row.Name,
		WebsiteUrl:        row.WebsiteUrl,
		Remark:            row.Remark,
		OpenaiBaseUrl:     row.OpenaiBaseUrl,
		AnthropicBaseUrl:  row.AnthropicBaseUrl,
		ApiKeySecretRef:   maskSecretRef(row.ApiKeySecretRef),
		Enabled:           row.Enabled,
		ModelCount:        modelCount,
		EnabledModelCount: enabledModelCount,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}
}

// modelToItem converts one model entity into a service projection.
func modelToItem(row *entity.Model) *ModelItem {
	if row == nil {
		return nil
	}
	return &ModelItem{
		Id:               row.Id,
		ProviderId:       row.ProviderId,
		CapabilityType:   row.CapabilityType,
		ModelName:        row.ModelName,
		Protocol:         row.Protocol,
		Source:           row.Source,
		SupportsThinking: row.SupportsThinking,
		SupportedEfforts: splitEfforts(row.SupportedEfforts),
		MaxInputTokens:   row.MaxInputTokens,
		MaxOutputTokens:  row.MaxOutputTokens,
		Enabled:          row.Enabled,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}

// invocationErrorCode extracts a stable business error code when available.
func invocationErrorCode(err error) string {
	if err == nil {
		return ""
	}
	if messageErr, ok := bizerr.As(err); ok {
		return messageErr.RuntimeCode()
	}
	return ""
}
