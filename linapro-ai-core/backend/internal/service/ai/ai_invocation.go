// This file implements masked invocation log writes and paged log queries.

package ai

import (
	"context"
	"time"

	"lina-core/pkg/plugin/capability/aicap/aitext"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-plugin-linapro-ai-core/backend/internal/dao"
	"lina-plugin-linapro-ai-core/backend/internal/model/do"
	"lina-plugin-linapro-ai-core/backend/internal/model/entity"
)

// invocationWriteInput carries the already-masked invocation fields that are
// safe to persist for both framework calls and management-side tier tests.
type invocationWriteInput struct {
	requestID        string
	capabilityType   string
	capabilityMethod string
	purpose          string
	tierCode         string
	sourcePluginID   string
	thinkingEffort   string
	binding          *resolvedTierBinding
	status           string
	usage            aitext.Usage
	latencyMs        int
	err              error
}

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

// CleanInvocations hard-deletes masked AI invocation logs within an optional
// creation time range. When no range is provided, all invocation logs are
// deleted in one bounded SQL statement.
func (s *serviceImpl) CleanInvocations(ctx context.Context, in InvocationCleanInput) (int, error) {
	if err := s.ensurePlatform(ctx); err != nil {
		return 0, err
	}
	cols := dao.Invocation.Columns()
	model := dao.Invocation.Ctx(ctx)
	hasFilter := false
	if in.StartedAt > 0 {
		model = model.WhereGTE(cols.CreatedAt, time.UnixMilli(in.StartedAt))
		hasFilter = true
	}
	if in.EndedAt > 0 {
		model = model.WhereLTE(cols.CreatedAt, time.UnixMilli(in.EndedAt))
		hasFilter = true
	}
	if !hasFilter {
		model = model.Where("1 = 1")
	}
	result, err := model.Delete()
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(affected), nil
}

// CleanupExpiredInvocations hard-deletes invocation logs older than the global retention boundary.
func (s *serviceImpl) CleanupExpiredInvocations(ctx context.Context, retentionDays int) (int, error) {
	if retentionDays <= 0 {
		return 0, nil
	}
	cols := dao.Invocation.Columns()
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	result, err := dao.Invocation.Ctx(ctx).
		WhereLT(cols.CreatedAt, cutoff).
		Delete()
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(affected), nil
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
	effort := ""
	if request.ThinkingEffort != nil {
		effort = string(*request.ThinkingEffort)
	} else if binding != nil {
		effort = binding.DefaultEffort
	}
	s.writeInvocationRecord(ctx, invocationWriteInput{
		requestID:        requestID,
		capabilityType:   string(request.CapabilityType()),
		capabilityMethod: string(request.CapabilityMethod()),
		purpose:          request.Purpose,
		tierCode:         string(request.Tier),
		sourcePluginID:   request.SourcePluginID,
		thinkingEffort:   effort,
		binding:          binding,
		status:           status,
		usage:            usage,
		latencyMs:        latencyMs,
		err:              err,
	})
}

// writeInvocationRecord stores one already-classified masked invocation record.
func (s *serviceImpl) writeInvocationRecord(ctx context.Context, input invocationWriteInput) {
	current := bizctxcap.CurrentFromContext(ctx)
	if s != nil && s.bizCtxSvc != nil {
		current = s.bizCtxSvc.Current(ctx)
	}
	providerID := int64(0)
	modelID := int64(0)
	providerName := ""
	modelName := ""
	protocol := ""
	if input.binding != nil {
		providerID = input.binding.ProviderId
		modelID = input.binding.ModelId
		providerName = input.binding.ProviderName
		modelName = input.binding.ModelName
		protocol = input.binding.Protocol
	}
	if _, insertErr := dao.Invocation.Ctx(ctx).Data(do.Invocation{
		RequestId:            input.requestID,
		CapabilityType:       input.capabilityType,
		CapabilityMethod:     input.capabilityMethod,
		Purpose:              input.purpose,
		TierCode:             input.tierCode,
		SourcePluginId:       input.sourcePluginID,
		TenantId:             current.TenantID,
		UserId:               current.UserID,
		ProviderId:           providerID,
		ModelId:              modelID,
		ProviderName:         providerName,
		ModelName:            modelName,
		Protocol:             protocol,
		ThinkingEffort:       input.thinkingEffort,
		Status:               input.status,
		InputTokens:          input.usage.InputTokens,
		OutputTokens:         input.usage.OutputTokens,
		LatencyMs:            input.latencyMs,
		AssetSummaryJson:     "{}",
		OperationSummaryJson: "{}",
		MetadataSummaryJson:  "{}",
		ErrorCode:            invocationErrorCode(input.err),
		ErrorSummary:         sanitizeErrorSummary(input.err),
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
