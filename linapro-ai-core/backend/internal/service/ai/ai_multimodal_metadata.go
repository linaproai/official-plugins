// This file implements multimodal Smart Center metadata management for
// provider endpoints, model method capabilities, and provider operation
// projections.

package ai

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-plugin-linapro-ai-core/backend/internal/dao"
	"lina-plugin-linapro-ai-core/backend/internal/model/do"
	"lina-plugin-linapro-ai-core/backend/internal/model/entity"
)

// ListProviderEndpoints returns provider protocol endpoints using bounded provider-scoped queries.
func (s *serviceImpl) ListProviderEndpoints(ctx context.Context, in ProviderEndpointListInput) ([]*ProviderEndpointItem, error) {
	if err := s.ensurePlatform(ctx); err != nil {
		return nil, err
	}
	if _, err := s.getProvider(ctx, in.ProviderId); err != nil {
		return nil, err
	}
	cols := dao.ProviderEndpoint.Columns()
	model := dao.ProviderEndpoint.Ctx(ctx).Where(do.ProviderEndpoint{ProviderId: in.ProviderId})
	if protocol := normalizeProtocol(in.Protocol); protocol != "" {
		model = model.Where(cols.Protocol, protocol)
	}
	if in.Enabled != nil {
		model = model.Where(cols.Enabled, normalizeEnabled(*in.Enabled))
	}
	rows := make([]*entity.ProviderEndpoint, 0)
	if err := model.OrderAsc(cols.Protocol).OrderAsc(cols.Id).Scan(&rows); err != nil {
		return nil, err
	}
	items := make([]*ProviderEndpointItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, providerEndpointToItem(row))
	}
	return items, nil
}

// CreateProviderEndpoint creates one provider protocol endpoint.
func (s *serviceImpl) CreateProviderEndpoint(ctx context.Context, in ProviderEndpointSaveInput) (int64, error) {
	if err := s.ensurePlatform(ctx); err != nil {
		return 0, err
	}
	if _, err := s.getProvider(ctx, in.ProviderId); err != nil {
		return 0, err
	}
	protocol := normalizeProtocol(in.Protocol)
	if protocol == "" || strings.TrimSpace(in.BaseUrl) == "" {
		return 0, bizerr.NewCode(CodeRequestInvalid)
	}
	id, err := dao.ProviderEndpoint.Ctx(ctx).Data(do.ProviderEndpoint{
		ProviderId:   in.ProviderId,
		Protocol:     protocol,
		BaseUrl:      strings.TrimSpace(in.BaseUrl),
		SecretRef:    strings.TrimSpace(in.SecretRef),
		Enabled:      normalizeEnabled(in.Enabled),
		MetadataJson: normalizeJSONText(in.MetadataJson),
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}
	return id, s.invalidateTierCache(ctx, "", "", "")
}

// UpdateProviderEndpoint updates one provider protocol endpoint.
func (s *serviceImpl) UpdateProviderEndpoint(ctx context.Context, in ProviderEndpointSaveInput) error {
	if err := s.ensurePlatform(ctx); err != nil {
		return err
	}
	row, err := s.getProviderEndpoint(ctx, in.ProviderId, in.Id)
	if err != nil {
		return err
	}
	protocol := normalizeProtocol(in.Protocol)
	if protocol == "" || strings.TrimSpace(in.BaseUrl) == "" {
		return bizerr.NewCode(CodeRequestInvalid)
	}
	secretRef := strings.TrimSpace(in.SecretRef)
	if shouldKeepExistingSecret(secretRef) {
		secretRef = row.SecretRef
	}
	_, err = dao.ProviderEndpoint.Ctx(ctx).
		Where(do.ProviderEndpoint{Id: in.Id, ProviderId: in.ProviderId}).
		Data(do.ProviderEndpoint{
			Protocol:     protocol,
			BaseUrl:      strings.TrimSpace(in.BaseUrl),
			SecretRef:    secretRef,
			Enabled:      normalizeEnabled(in.Enabled),
			MetadataJson: normalizeJSONText(in.MetadataJson),
		}).
		Update()
	if err != nil {
		return err
	}
	return s.invalidateTierCache(ctx, "", "", "")
}

// DeleteProviderEndpoint soft-deletes one endpoint after reference checks.
func (s *serviceImpl) DeleteProviderEndpoint(ctx context.Context, providerID int64, id int64) error {
	if err := s.ensurePlatform(ctx); err != nil {
		return err
	}
	if _, err := s.getProviderEndpoint(ctx, providerID, id); err != nil {
		return err
	}
	inUse, err := s.providerEndpointReferenced(ctx, id)
	if err != nil {
		return err
	}
	if inUse {
		return bizerr.NewCode(CodeProviderEndpointInUse)
	}
	if _, err = dao.ProviderEndpoint.Ctx(ctx).Where(do.ProviderEndpoint{Id: id, ProviderId: providerID}).Delete(); err != nil {
		return err
	}
	return s.invalidateTierCache(ctx, "", "", "")
}

// syncProviderFormEndpoints applies the provider drawer's fixed OpenAI and
// Anthropic endpoint payload inside the caller's provider transaction.
func (s *serviceImpl) syncProviderFormEndpoints(ctx context.Context, providerID int64, inputs []ProviderEndpointSaveInput) error {
	if len(inputs) == 0 {
		return nil
	}
	seenProtocols := make(map[string]struct{}, len(inputs))
	seenIDs := make(map[int64]struct{}, len(inputs))
	for _, input := range inputs {
		protocol := normalizeProviderFormProtocol(input.Protocol)
		if protocol == "" {
			return bizerr.NewCode(CodeRequestInvalid)
		}
		if _, ok := seenProtocols[protocol]; ok {
			return bizerr.NewCode(CodeRequestInvalid)
		}
		seenProtocols[protocol] = struct{}{}
		if input.Id > 0 {
			if _, ok := seenIDs[input.Id]; ok {
				return bizerr.NewCode(CodeRequestInvalid)
			}
			seenIDs[input.Id] = struct{}{}
		}
		if err := s.saveProviderFormEndpoint(ctx, providerID, protocol, input); err != nil {
			return err
		}
	}
	return nil
}

func (s *serviceImpl) saveProviderFormEndpoint(ctx context.Context, providerID int64, protocol string, in ProviderEndpointSaveInput) error {
	baseURL := strings.TrimSpace(in.BaseUrl)
	if in.Id <= 0 && baseURL == "" && strings.TrimSpace(in.SecretRef) == "" && endpointMetadataEmpty(in.MetadataJson) {
		return nil
	}
	if in.Id <= 0 {
		if baseURL == "" {
			return bizerr.NewCode(CodeRequestInvalid)
		}
		_, err := dao.ProviderEndpoint.Ctx(ctx).Data(do.ProviderEndpoint{
			ProviderId:   providerID,
			Protocol:     protocol,
			BaseUrl:      baseURL,
			SecretRef:    strings.TrimSpace(in.SecretRef),
			Enabled:      normalizeEnabled(in.Enabled),
			MetadataJson: normalizeJSONText(in.MetadataJson),
		}).Insert()
		return err
	}
	row, err := s.getProviderEndpoint(ctx, providerID, in.Id)
	if err != nil {
		return err
	}
	if normalizeProviderFormProtocol(row.Protocol) == "" {
		return bizerr.NewCode(CodeRequestInvalid)
	}
	if baseURL == "" {
		return s.deleteProviderEndpointInTransaction(ctx, providerID, in.Id)
	}
	if row.Protocol != protocol {
		inUse, err := s.providerEndpointReferenced(ctx, in.Id)
		if err != nil {
			return err
		}
		if inUse {
			return bizerr.NewCode(CodeProviderEndpointInUse)
		}
	}
	secretRef := strings.TrimSpace(in.SecretRef)
	if shouldKeepExistingSecret(secretRef) {
		secretRef = row.SecretRef
	}
	_, err = dao.ProviderEndpoint.Ctx(ctx).
		Where(do.ProviderEndpoint{Id: in.Id, ProviderId: providerID}).
		Data(do.ProviderEndpoint{
			Protocol:     protocol,
			BaseUrl:      baseURL,
			SecretRef:    secretRef,
			Enabled:      normalizeEnabled(in.Enabled),
			MetadataJson: normalizeJSONText(in.MetadataJson),
		}).
		Update()
	return err
}

func (s *serviceImpl) deleteProviderEndpointInTransaction(ctx context.Context, providerID int64, id int64) error {
	inUse, err := s.providerEndpointReferenced(ctx, id)
	if err != nil {
		return err
	}
	if inUse {
		return bizerr.NewCode(CodeProviderEndpointInUse)
	}
	_, err = dao.ProviderEndpoint.Ctx(ctx).Where(do.ProviderEndpoint{Id: id, ProviderId: providerID}).Delete()
	return err
}

func normalizeProviderFormProtocol(value string) string {
	switch normalizeProtocol(value) {
	case protocolOpenAI:
		return protocolOpenAI
	case protocolAnthropic:
		return protocolAnthropic
	default:
		return ""
	}
}

func endpointMetadataEmpty(value string) bool {
	trimmed := strings.TrimSpace(value)
	return trimmed == "" || trimmed == "{}"
}

// ListModelCapabilities returns explicit method capability declarations for one model.
func (s *serviceImpl) ListModelCapabilities(ctx context.Context, modelID int64) ([]*ModelCapabilityItem, error) {
	if err := s.ensurePlatform(ctx); err != nil {
		return nil, err
	}
	if _, err := s.getModel(ctx, modelID); err != nil {
		return nil, err
	}
	cols := dao.ModelCapability.Columns()
	rows := make([]*entity.ModelCapability, 0)
	if err := dao.ModelCapability.Ctx(ctx).
		Where(do.ModelCapability{ModelId: modelID}).
		OrderAsc(cols.CapabilityType).
		OrderAsc(cols.CapabilityMethod).
		Scan(&rows); err != nil {
		return nil, err
	}
	items := make([]*ModelCapabilityItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, modelCapabilityToItem(row))
	}
	return items, nil
}

// UpsertModelCapabilities replaces explicit method capability declarations for one model.
func (s *serviceImpl) UpsertModelCapabilities(ctx context.Context, modelID int64, items []ModelCapabilitySaveInput) error {
	if err := s.ensurePlatform(ctx); err != nil {
		return err
	}
	model, err := s.getModel(ctx, modelID)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return bizerr.NewCode(CodeRequestInvalid)
	}
	keepKeys, err := modelCapabilityInputKeys(modelID, items)
	if err != nil {
		return err
	}
	err = dao.ModelCapability.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		existingRows, err := s.modelCapabilitiesByModel(ctx, []int64{modelID})
		if err != nil {
			return err
		}
		removedIDs, removedKeys := removedModelCapabilityIDs(existingRows, keepKeys)
		if len(removedKeys) > 0 {
			referenced, err := s.modelCapabilityKeysReferencedByTiers(ctx, modelID, removedKeys)
			if err != nil {
				return err
			}
			if referenced {
				return bizerr.NewCode(CodeModelInUse)
			}
		}
		for _, item := range items {
			if err := s.upsertModelCapability(ctx, model, item); err != nil {
				return err
			}
		}
		if len(removedIDs) > 0 {
			if _, err := dao.ModelCapability.Ctx(ctx).
				WhereIn(dao.ModelCapability.Columns().Id, removedIDs).
				Delete(); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return s.invalidateTierCache(ctx, "", "", "")
}

// modelCapabilityInputKeys returns normalized capability identities from the save payload.
func modelCapabilityInputKeys(modelID int64, items []ModelCapabilitySaveInput) (map[string]struct{}, error) {
	keys := make(map[string]struct{}, len(items))
	for _, item := range items {
		capabilityType := normalizeCapabilityType(item.CapabilityType)
		capabilityMethod := normalizeCapabilityMethod(item.CapabilityMethod)
		if capabilityType == "" || capabilityMethod == "" {
			return nil, bizerr.NewCode(CodeRequestInvalid)
		}
		key := modelCapabilityKey(modelID, capabilityType, capabilityMethod)
		if _, exists := keys[key]; exists {
			return nil, bizerr.NewCode(CodeRequestInvalid)
		}
		keys[key] = struct{}{}
	}
	return keys, nil
}

// removedModelCapabilityIDs returns existing capability rows absent from the replacement payload.
func removedModelCapabilityIDs(rows []*entity.ModelCapability, keepKeys map[string]struct{}) ([]int64, map[string]struct{}) {
	ids := make([]int64, 0)
	keys := make(map[string]struct{})
	for _, row := range rows {
		if row == nil || row.Id <= 0 {
			continue
		}
		key := modelCapabilityKey(row.ModelId, row.CapabilityType, row.CapabilityMethod)
		if _, keep := keepKeys[key]; keep {
			continue
		}
		ids = append(ids, row.Id)
		keys[key] = struct{}{}
	}
	return ids, keys
}

// modelCapabilityKeysReferencedByTiers reports whether removed capability keys are still tier-bound.
func (s *serviceImpl) modelCapabilityKeysReferencedByTiers(ctx context.Context, modelID int64, keys map[string]struct{}) (bool, error) {
	if modelID <= 0 || len(keys) == 0 {
		return false, nil
	}
	bindingRows := make([]*entity.TierBinding, 0)
	if err := dao.TierBinding.Ctx(ctx).
		Fields(dao.TierBinding.Columns().TierId).
		Where(do.TierBinding{ModelId: modelID}).
		Scan(&bindingRows); err != nil {
		return false, err
	}
	tierIDs := make([]int64, 0, len(bindingRows))
	seenTierIDs := make(map[int64]struct{}, len(bindingRows))
	for _, row := range bindingRows {
		if row == nil || row.TierId <= 0 {
			continue
		}
		if _, seen := seenTierIDs[row.TierId]; seen {
			continue
		}
		seenTierIDs[row.TierId] = struct{}{}
		tierIDs = append(tierIDs, row.TierId)
	}
	tiers, err := s.tiersByID(ctx, tierIDs)
	if err != nil {
		return false, err
	}
	for _, row := range bindingRows {
		if row == nil {
			continue
		}
		tier := tiers[row.TierId]
		if tier == nil {
			continue
		}
		if _, referenced := keys[modelCapabilityKey(modelID, tier.CapabilityType, tier.CapabilityMethod)]; referenced {
			return true, nil
		}
	}
	return false, nil
}

// ListProviderOperations returns masked provider operation projections with database-side filters.
func (s *serviceImpl) ListProviderOperations(ctx context.Context, in ProviderOperationListInput) (*ProviderOperationListOutput, error) {
	if err := s.ensurePlatform(ctx); err != nil {
		return nil, err
	}
	pageNum, pageSize := normalizePage(in.PageNum, in.PageSize)
	cols := dao.ProviderOperation.Columns()
	model := dao.ProviderOperation.Ctx(ctx)
	if in.CapabilityType != "" {
		model = model.Where(cols.CapabilityType, normalizeCapabilityType(in.CapabilityType))
	}
	if in.CapabilityMethod != "" {
		model = model.Where(cols.CapabilityMethod, normalizeCapabilityMethod(in.CapabilityMethod))
	}
	if in.Purpose != "" {
		model = model.Where(cols.Purpose, strings.TrimSpace(in.Purpose))
	}
	if in.Status != "" {
		model = model.Where(cols.Status, strings.TrimSpace(in.Status))
	}
	if in.ProviderId > 0 {
		model = model.Where(cols.ProviderId, in.ProviderId)
	}
	if in.ModelId > 0 {
		model = model.Where(cols.ModelId, in.ModelId)
	}
	if in.SourcePluginId != "" {
		model = model.Where(cols.SourcePluginId, strings.TrimSpace(in.SourcePluginId))
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
	rows := make([]*entity.ProviderOperation, 0)
	if err = model.Page(pageNum, pageSize).OrderDesc(cols.CreatedAt).Scan(&rows); err != nil {
		return nil, err
	}
	items := make([]*ProviderOperationItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, providerOperationToItem(row))
	}
	return &ProviderOperationListOutput{List: items, Total: total}, nil
}

// GetProviderOperation returns one provider operation projection by opaque reference.
func (s *serviceImpl) GetProviderOperation(ctx context.Context, operationRef string) (*ProviderOperationItem, error) {
	if err := s.ensurePlatform(ctx); err != nil {
		return nil, err
	}
	ref := strings.TrimSpace(operationRef)
	if ref == "" {
		return nil, bizerr.NewCode(CodeProviderOperationNotFound)
	}
	var row *entity.ProviderOperation
	if err := dao.ProviderOperation.Ctx(ctx).Where(do.ProviderOperation{OperationRef: ref}).Scan(&row); err != nil {
		return nil, err
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeProviderOperationNotFound)
	}
	return providerOperationToItem(row), nil
}

func (s *serviceImpl) upsertModelCapability(ctx context.Context, model *entity.Model, item ModelCapabilitySaveInput) error {
	capabilityType := normalizeCapabilityType(item.CapabilityType)
	capabilityMethod := normalizeCapabilityMethod(item.CapabilityMethod)
	if capabilityType == "" || capabilityMethod == "" {
		return bizerr.NewCode(CodeRequestInvalid)
	}
	supportedEfforts, _, err := normalizeEfforts(item.SupportedEfforts)
	if err != nil {
		return err
	}
	if item.EndpointId > 0 {
		if _, err := s.requireProviderEndpoint(ctx, model.ProviderId, item.EndpointId, model.Protocol); err != nil {
			return err
		}
	}
	data := do.ModelCapability{
		ModelId:           model.Id,
		EndpointId:        item.EndpointId,
		CapabilityType:    capabilityType,
		CapabilityMethod:  capabilityMethod,
		InputModalities:   joinCSV(item.InputModalities),
		OutputModalities:  joinCSV(item.OutputModalities),
		MaxInputTokens:    item.MaxInputTokens,
		MaxOutputTokens:   item.MaxOutputTokens,
		MaxInputAssets:    item.MaxInputAssets,
		MaxOutputAssets:   item.MaxOutputAssets,
		MaxAssetBytes:     item.MaxAssetBytes,
		SupportsStreaming: normalizeEnabled(item.SupportsStreaming),
		SupportsOperation: normalizeEnabled(item.SupportsOperation),
		SupportsThinking:  normalizeEnabled(item.SupportsThinking),
		SupportedEfforts:  supportedEfforts,
		Enabled:           normalizeEnabled(item.Enabled),
	}
	var existing *entity.ModelCapability
	if err := dao.ModelCapability.Ctx(ctx).
		Where(do.ModelCapability{ModelId: model.Id, CapabilityType: capabilityType, CapabilityMethod: capabilityMethod}).
		Scan(&existing); err != nil {
		return err
	}
	if existing == nil {
		_, err := dao.ModelCapability.Ctx(ctx).Data(data).Insert()
		return err
	}
	_, err = dao.ModelCapability.Ctx(ctx).Where(do.ModelCapability{Id: existing.Id}).Data(data).Update()
	return err
}

func (s *serviceImpl) getProviderEndpoint(ctx context.Context, providerID int64, id int64) (*entity.ProviderEndpoint, error) {
	if providerID <= 0 || id <= 0 {
		return nil, bizerr.NewCode(CodeProviderEndpointNotFound)
	}
	var row *entity.ProviderEndpoint
	if err := dao.ProviderEndpoint.Ctx(ctx).Where(do.ProviderEndpoint{Id: id, ProviderId: providerID}).Scan(&row); err != nil {
		return nil, err
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeProviderEndpointNotFound)
	}
	return row, nil
}

func (s *serviceImpl) requireProviderEndpoint(ctx context.Context, providerID int64, id int64, protocol string) (*entity.ProviderEndpoint, error) {
	protocol = normalizeProtocol(protocol)
	if providerID <= 0 || id <= 0 || protocol == "" {
		return nil, bizerr.NewCode(CodeProviderProtocolRequired)
	}
	row, err := s.getProviderEndpoint(ctx, providerID, id)
	if err != nil {
		return nil, err
	}
	if row.Enabled != enabledYes || row.Protocol != protocol {
		return nil, bizerr.NewCode(CodeProviderProtocolRequired)
	}
	return row, nil
}

func (s *serviceImpl) enabledSyncEndpoints(ctx context.Context, providerID int64, protocol string) ([]*entity.ProviderEndpoint, error) {
	if providerID <= 0 {
		return nil, bizerr.NewCode(CodeProviderProtocolRequired)
	}
	cols := dao.ProviderEndpoint.Columns()
	model := dao.ProviderEndpoint.Ctx(ctx).Where(do.ProviderEndpoint{
		ProviderId: providerID,
		Enabled:    enabledYes,
	})
	if protocol != "" {
		model = model.Where(cols.Protocol, protocol)
	}
	rows := make([]*entity.ProviderEndpoint, 0)
	if err := model.OrderAsc(cols.Id).Scan(&rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, bizerr.NewCode(CodeProviderProtocolRequired)
	}
	return rows, nil
}

func (s *serviceImpl) providerEndpointReferenced(ctx context.Context, endpointID int64) (bool, error) {
	modelCount, err := dao.Model.Ctx(ctx).Where(do.Model{EndpointId: endpointID}).Count()
	return modelCount > 0, err
}

func providerEndpointToItem(row *entity.ProviderEndpoint) *ProviderEndpointItem {
	if row == nil {
		return nil
	}
	return &ProviderEndpointItem{
		Id:           row.Id,
		ProviderId:   row.ProviderId,
		Protocol:     row.Protocol,
		BaseUrl:      row.BaseUrl,
		SecretRef:    maskSecretRef(row.SecretRef),
		Enabled:      row.Enabled,
		MetadataJson: row.MetadataJson,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

// countEndpointsByProvider counts all and enabled endpoints using one batch query.
func (s *serviceImpl) countEndpointsByProvider(ctx context.Context, providerIDs []int64) (map[int64]int, map[int64]int, error) {
	counts := make(map[int64]int, len(providerIDs))
	enabledCounts := make(map[int64]int, len(providerIDs))
	if len(providerIDs) == 0 {
		return counts, enabledCounts, nil
	}
	cols := dao.ProviderEndpoint.Columns()
	type endpointCountRow struct {
		ProviderId           int64 `orm:"provider_id"`
		EndpointCount        int64 `orm:"endpoint_count"`
		EnabledEndpointCount int64 `orm:"enabled_endpoint_count"`
	}
	rows := make([]endpointCountRow, 0)
	if err := dao.ProviderEndpoint.Ctx(ctx).
		Fields(cols.ProviderId, "COUNT(*) AS endpoint_count", "SUM(CASE WHEN "+cols.Enabled+" = 1 THEN 1 ELSE 0 END) AS enabled_endpoint_count").
		WhereIn(cols.ProviderId, providerIDs).
		Group(cols.ProviderId).
		Scan(&rows); err != nil {
		return nil, nil, err
	}
	for _, row := range rows {
		counts[row.ProviderId] = int(row.EndpointCount)
		enabledCounts[row.ProviderId] = int(row.EnabledEndpointCount)
	}
	return counts, enabledCounts, nil
}

// listEndpointSummariesByProvider returns compact endpoint summaries for the current provider page.
func (s *serviceImpl) listEndpointSummariesByProvider(ctx context.Context, providerIDs []int64) (map[int64][]*ProviderEndpointItem, error) {
	result := make(map[int64][]*ProviderEndpointItem, len(providerIDs))
	if len(providerIDs) == 0 {
		return result, nil
	}
	cols := dao.ProviderEndpoint.Columns()
	rows := make([]*entity.ProviderEndpoint, 0)
	if err := dao.ProviderEndpoint.Ctx(ctx).
		Fields(cols.Id, cols.ProviderId, cols.Protocol, cols.BaseUrl, cols.SecretRef, cols.Enabled, cols.MetadataJson, cols.CreatedAt, cols.UpdatedAt).
		WhereIn(cols.ProviderId, providerIDs).
		OrderAsc(cols.ProviderId).
		OrderAsc(cols.Protocol).
		OrderAsc(cols.Id).
		Scan(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		result[row.ProviderId] = append(result[row.ProviderId], providerEndpointToItem(row))
	}
	return result, nil
}

func modelCapabilityToItem(row *entity.ModelCapability) *ModelCapabilityItem {
	if row == nil {
		return nil
	}
	return &ModelCapabilityItem{
		Id:                row.Id,
		ModelId:           row.ModelId,
		EndpointId:        row.EndpointId,
		CapabilityType:    row.CapabilityType,
		CapabilityMethod:  row.CapabilityMethod,
		InputModalities:   splitCSV(row.InputModalities),
		OutputModalities:  splitCSV(row.OutputModalities),
		MaxInputTokens:    row.MaxInputTokens,
		MaxOutputTokens:   row.MaxOutputTokens,
		MaxInputAssets:    row.MaxInputAssets,
		MaxOutputAssets:   row.MaxOutputAssets,
		MaxAssetBytes:     row.MaxAssetBytes,
		SupportsStreaming: row.SupportsStreaming,
		SupportsOperation: row.SupportsOperation,
		SupportsThinking:  row.SupportsThinking,
		SupportedEfforts:  splitEfforts(row.SupportedEfforts),
		Enabled:           row.Enabled,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}
}

func providerOperationToItem(row *entity.ProviderOperation) *ProviderOperationItem {
	if row == nil {
		return nil
	}
	return &ProviderOperationItem{
		Id:               row.Id,
		OperationRef:     row.OperationRef,
		CapabilityType:   row.CapabilityType,
		CapabilityMethod: row.CapabilityMethod,
		Purpose:          row.Purpose,
		SourcePluginId:   row.SourcePluginId,
		ProviderId:       row.ProviderId,
		ModelId:          row.ModelId,
		ProviderName:     row.ProviderName,
		ModelName:        row.ModelName,
		Protocol:         row.Protocol,
		Status:           row.Status,
		NextPollAfterMs:  row.NextPollAfterMs,
		ExpiresAt:        row.ExpiresAt,
		AssetSummaryJson: row.AssetSummaryJson,
		ErrorCode:        row.ErrorCode,
		ErrorSummary:     sanitizeText(row.ErrorSummary, 512),
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}

func normalizeJSONText(value string) string {
	if strings.TrimSpace(value) == "" {
		return "{}"
	}
	return strings.TrimSpace(value)
}

func joinCSV(values []string) string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return strings.Join(out, ",")
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
