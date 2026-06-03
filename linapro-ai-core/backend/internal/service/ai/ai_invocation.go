// This file implements masked invocation log writes and paged log queries.

package ai

import (
	"context"
	"time"

	"lina-core/pkg/plugin/capability/ai/aitext"
	"lina-core/pkg/plugin/capability/contract"
	"lina-plugin-linapro-ai-core/backend/internal/dao"
	"lina-plugin-linapro-ai-core/backend/internal/model/do"
	"lina-plugin-linapro-ai-core/backend/internal/model/entity"
)

// ListInvocations returns masked AI invocation logs with database-side filters.
func (s *serviceImpl) ListInvocations(ctx context.Context, in InvocationListInput) (*InvocationListOutput, error) {
	if err := s.ensurePlatform(ctx); err != nil {
		return nil, err
	}
	pageNum, pageSize := normalizePage(in.PageNum, in.PageSize)
	cols := dao.Invocation.Columns()
	model := dao.Invocation.Ctx(ctx)
	if in.CapabilityType != "" {
		model = model.Where(cols.CapabilityType, normalizeCapabilityType(in.CapabilityType))
	}
	if in.CapabilityMethod != "" {
		model = model.Where(cols.CapabilityMethod, normalizeCapabilityMethod(in.CapabilityMethod))
	}
	if in.Purpose != "" {
		model = model.Where(cols.Purpose, in.Purpose)
	}
	if in.TierCode != "" {
		model = model.Where(cols.TierCode, normalizeTierCode(in.TierCode))
	}
	if in.Status != "" {
		model = model.Where(cols.Status, in.Status)
	}
	if in.ProviderId > 0 {
		model = model.Where(cols.ProviderId, in.ProviderId)
	}
	if in.ModelId > 0 {
		model = model.Where(cols.ModelId, in.ModelId)
	}
	if in.SourcePluginId != "" {
		model = model.Where(cols.SourcePluginId, in.SourcePluginId)
	}
	if in.StartedAt > 0 {
		model = model.WhereGTE(cols.CreatedAt, time.UnixMilli(in.StartedAt))
	}
	if in.EndedAt > 0 {
		model = model.WhereLTE(cols.CreatedAt, time.UnixMilli(in.EndedAt))
	}
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	rows := make([]*entity.Invocation, 0)
	if err = model.Page(pageNum, pageSize).OrderDesc(cols.CreatedAt).Scan(&rows); err != nil {
		return nil, err
	}
	items := make([]*InvocationItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, invocationToItem(row))
	}
	return &InvocationListOutput{List: items, Total: total}, nil
}

// writeInvocation stores one masked invocation record. It intentionally avoids
// prompt, response, request body, response body, and secret content.
func (s *serviceImpl) writeInvocation(
	ctx context.Context,
	requestID string,
	request aitext.ProviderRequest,
	binding *resolvedTierBinding,
	status string,
	usage aitext.Usage,
	latencyMs int,
	err error,
) {
	current := contract.CurrentFromContext(ctx)
	if s != nil && s.bizCtxSvc != nil {
		current = s.bizCtxSvc.Current(ctx)
	}
	providerID := int64(0)
	modelID := int64(0)
	providerName := ""
	modelName := ""
	protocol := ""
	if binding != nil {
		providerID = binding.ProviderId
		modelID = binding.ModelId
		providerName = binding.ProviderName
		modelName = binding.ModelName
		protocol = binding.Protocol
	}
	effort := ""
	if request.ThinkingEffort != nil {
		effort = string(*request.ThinkingEffort)
	} else if binding != nil {
		effort = binding.DefaultEffort
	}
	if _, insertErr := dao.Invocation.Ctx(ctx).Data(do.Invocation{
		RequestId:            requestID,
		CapabilityType:       string(request.CapabilityType()),
		CapabilityMethod:     string(request.CapabilityMethod()),
		Purpose:              request.Purpose,
		TierCode:             string(request.Tier),
		SourcePluginId:       request.SourcePluginID,
		TenantId:             current.TenantID,
		UserId:               current.UserID,
		ProviderId:           providerID,
		ModelId:              modelID,
		ProviderName:         providerName,
		ModelName:            modelName,
		Protocol:             protocol,
		ThinkingEffort:       effort,
		Status:               status,
		InputTokens:          usage.InputTokens,
		OutputTokens:         usage.OutputTokens,
		LatencyMs:            latencyMs,
		AssetSummaryJson:     "{}",
		OperationSummaryJson: "{}",
		MetadataSummaryJson:  "{}",
		ErrorCode:            invocationErrorCode(err),
		ErrorSummary:         sanitizeErrorSummary(err),
	}).Insert(); insertErr != nil {
		// Invocation logging is diagnostic and must not replace the provider error.
		return
	}
}

// invocationToItem converts one invocation entity into a masked service projection.
func invocationToItem(row *entity.Invocation) *InvocationItem {
	if row == nil {
		return nil
	}
	return &InvocationItem{
		Id:                   row.Id,
		RequestId:            row.RequestId,
		CapabilityType:       row.CapabilityType,
		CapabilityMethod:     row.CapabilityMethod,
		Purpose:              row.Purpose,
		TierCode:             row.TierCode,
		SourcePluginId:       row.SourcePluginId,
		TenantId:             row.TenantId,
		UserId:               row.UserId,
		ProviderId:           row.ProviderId,
		ModelId:              row.ModelId,
		ProviderName:         row.ProviderName,
		ModelName:            row.ModelName,
		Protocol:             row.Protocol,
		ThinkingEffort:       row.ThinkingEffort,
		Status:               row.Status,
		InputTokens:          row.InputTokens,
		OutputTokens:         row.OutputTokens,
		LatencyMs:            row.LatencyMs,
		AssetSummaryJson:     row.AssetSummaryJson,
		OperationSummaryJson: row.OperationSummaryJson,
		MetadataSummaryJson:  row.MetadataSummaryJson,
		ErrorCode:            row.ErrorCode,
		ErrorSummary:         row.ErrorSummary,
		CreatedAt:            row.CreatedAt,
	}
}
