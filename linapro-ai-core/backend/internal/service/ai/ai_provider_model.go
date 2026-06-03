// This file implements provider and model management with batched model count
// assembly and tier-reference protection.

package ai

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-plugin-linapro-ai-core/backend/internal/dao"
	"lina-plugin-linapro-ai-core/backend/internal/model/do"
	"lina-plugin-linapro-ai-core/backend/internal/model/entity"
)

// ListProviders returns a platform-only paged provider list with batched model summaries.
func (s *serviceImpl) ListProviders(ctx context.Context, in ProviderListInput) (*ProviderListOutput, error) {
	if err := s.ensurePlatform(ctx); err != nil {
		return nil, err
	}
	pageNum, pageSize := normalizePage(in.PageNum, in.PageSize)
	cols := dao.Provider.Columns()
	model := dao.Provider.Ctx(ctx)
	if keyword := strings.TrimSpace(in.Keyword); keyword != "" {
		model = model.WhereLike(cols.Name, "%"+keyword+"%")
	}
	if in.Enabled != nil {
		model = model.Where(cols.Enabled, normalizeEnabled(*in.Enabled))
	}
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	rows := make([]*entity.Provider, 0)
	if err = model.Page(pageNum, pageSize).OrderDesc(cols.Id).Scan(&rows); err != nil {
		return nil, err
	}
	providerIDs := collectProviderIDs(rows)
	counts, enabledCounts, err := s.countModelsByProvider(ctx, providerIDs)
	if err != nil {
		return nil, err
	}
	modelsByProvider, err := s.listModelSummariesByProvider(ctx, providerIDs)
	if err != nil {
		return nil, err
	}
	endpointCounts, enabledEndpointCounts, err := s.countEndpointsByProvider(ctx, providerIDs)
	if err != nil {
		return nil, err
	}
	endpointsByProvider, err := s.listEndpointSummariesByProvider(ctx, providerIDs)
	if err != nil {
		return nil, err
	}
	items := make([]*ProviderItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, providerToItem(
			row,
			counts[row.Id],
			enabledCounts[row.Id],
			modelsByProvider[row.Id],
			endpointCounts[row.Id],
			enabledEndpointCounts[row.Id],
			endpointsByProvider[row.Id],
		))
	}
	return &ProviderListOutput{List: items, Total: total}, nil
}

// GetProvider returns one provider projection with model counts.
func (s *serviceImpl) GetProvider(ctx context.Context, id int64) (*ProviderItem, error) {
	if err := s.ensurePlatform(ctx); err != nil {
		return nil, err
	}
	row, err := s.getProvider(ctx, id)
	if err != nil {
		return nil, err
	}
	counts, enabledCounts, err := s.countModelsByProvider(ctx, []int64{id})
	if err != nil {
		return nil, err
	}
	modelsByProvider, err := s.listModelSummariesByProvider(ctx, []int64{id})
	if err != nil {
		return nil, err
	}
	endpointCounts, enabledEndpointCounts, err := s.countEndpointsByProvider(ctx, []int64{id})
	if err != nil {
		return nil, err
	}
	endpointsByProvider, err := s.listEndpointSummariesByProvider(ctx, []int64{id})
	if err != nil {
		return nil, err
	}
	return providerToItem(
		row,
		counts[id],
		enabledCounts[id],
		modelsByProvider[id],
		endpointCounts[id],
		enabledEndpointCounts[id],
		endpointsByProvider[id],
	), nil
}

// CreateProvider creates one provider.
func (s *serviceImpl) CreateProvider(ctx context.Context, in ProviderSaveInput) (int64, error) {
	if err := s.ensurePlatform(ctx); err != nil {
		return 0, err
	}
	if strings.TrimSpace(in.Name) == "" {
		return 0, bizerr.NewCode(CodeRequestInvalid)
	}
	id, err := dao.Provider.Ctx(ctx).Data(do.Provider{
		Name:       strings.TrimSpace(in.Name),
		WebsiteUrl: strings.TrimSpace(in.WebsiteUrl),
		Remark:     strings.TrimSpace(in.Remark),
		Enabled:    normalizeEnabled(in.Enabled),
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}
	if err = s.InvalidateTierCache(ctx, "", "", ""); err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateProvider updates one provider.
func (s *serviceImpl) UpdateProvider(ctx context.Context, in ProviderSaveInput) error {
	if err := s.ensurePlatform(ctx); err != nil {
		return err
	}
	_, err := s.getProvider(ctx, in.Id)
	if err != nil {
		return err
	}
	if strings.TrimSpace(in.Name) == "" {
		return bizerr.NewCode(CodeRequestInvalid)
	}
	_, err = dao.Provider.Ctx(ctx).
		Where(do.Provider{Id: in.Id}).
		Data(do.Provider{
			Name:       strings.TrimSpace(in.Name),
			WebsiteUrl: strings.TrimSpace(in.WebsiteUrl),
			Remark:     strings.TrimSpace(in.Remark),
			Enabled:    normalizeEnabled(in.Enabled),
		}).
		Update()
	if err != nil {
		return err
	}
	return s.InvalidateTierCache(ctx, "", "", "")
}

// DeleteProvider soft-deletes one provider and its models after reference checks.
func (s *serviceImpl) DeleteProvider(ctx context.Context, id int64) error {
	if err := s.ensurePlatform(ctx); err != nil {
		return err
	}
	if _, err := s.getProvider(ctx, id); err != nil {
		return err
	}
	inUse, err := s.providerReferenced(ctx, id)
	if err != nil {
		return err
	}
	if inUse {
		return bizerr.NewCode(CodeProviderInUse)
	}
	err = dao.Provider.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		modelRows := make([]*entity.Model, 0)
		if err := dao.Model.Ctx(ctx).Fields(dao.Model.Columns().Id).Where(do.Model{ProviderId: id}).Scan(&modelRows); err != nil {
			return err
		}
		modelIDs := collectModelIDs(modelRows)
		if len(modelIDs) > 0 {
			if _, err := dao.ModelCapability.Ctx(ctx).
				WhereIn(dao.ModelCapability.Columns().ModelId, modelIDs).
				Delete(); err != nil {
				return err
			}
		}
		if _, err := dao.Model.Ctx(ctx).Where(do.Model{ProviderId: id}).Delete(); err != nil {
			return err
		}
		if _, err := dao.ProviderEndpoint.Ctx(ctx).Where(do.ProviderEndpoint{ProviderId: id}).Delete(); err != nil {
			return err
		}
		_, err := dao.Provider.Ctx(ctx).Where(do.Provider{Id: id}).Delete()
		return err
	})
	if err != nil {
		return err
	}
	return s.InvalidateTierCache(ctx, "", "", "")
}

// ListModels returns one bounded provider-owned model page.
func (s *serviceImpl) ListModels(ctx context.Context, in ModelListInput) (*ModelListOutput, error) {
	if err := s.ensurePlatform(ctx); err != nil {
		return nil, err
	}
	if _, err := s.getProvider(ctx, in.ProviderId); err != nil {
		return nil, err
	}
	pageNum, pageSize := normalizePage(in.PageNum, in.PageSize)
	cols := dao.Model.Columns()
	model := dao.Model.Ctx(ctx).Where(cols.ProviderId, in.ProviderId)
	capabilityType := normalizeCapabilityType(in.CapabilityType)
	capabilityMethod := normalizeCapabilityMethod(in.CapabilityMethod)
	capCols := dao.ModelCapability.Columns()
	capabilityQuery := dao.ModelCapability.Ctx(ctx).
		Fields(capCols.ModelId).
		Where(dao.ModelCapability.Table() + "." + capCols.ModelId + " = " + dao.Model.Table() + "." + cols.Id).
		Where(do.ModelCapability{CapabilityType: capabilityType, CapabilityMethod: capabilityMethod})
	if in.Enabled != nil {
		enabled := normalizeEnabled(*in.Enabled)
		model = model.Where(cols.Enabled, enabled)
		capabilityQuery = capabilityQuery.Where(capCols.Enabled, enabled)
	}
	model = model.Where("EXISTS ?", capabilityQuery)
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	rows := make([]*entity.Model, 0)
	if err = model.Page(pageNum, pageSize).OrderAsc(cols.Id).Scan(&rows); err != nil {
		return nil, err
	}
	capabilities, err := s.modelCapabilitiesByModelMethod(ctx, collectModelIDs(rows), capabilityType, capabilityMethod)
	if err != nil {
		return nil, err
	}
	items := make([]*ModelItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, modelToItem(row, capabilities[row.Id]))
	}
	return &ModelListOutput{List: items, Total: total}, nil
}

// CreateModel creates one provider-owned model.
func (s *serviceImpl) CreateModel(ctx context.Context, in ModelSaveInput) (int64, error) {
	if err := s.ensurePlatform(ctx); err != nil {
		return 0, err
	}
	if _, err := s.getProvider(ctx, in.ProviderId); err != nil {
		return 0, err
	}
	protocol := normalizeProtocol(in.Protocol)
	if protocol == "" || strings.TrimSpace(in.ModelName) == "" {
		return 0, bizerr.NewCode(CodeRequestInvalid)
	}
	if _, err := s.requireProviderEndpoint(ctx, in.ProviderId, in.EndpointId, protocol); err != nil {
		return 0, err
	}
	source := strings.TrimSpace(in.Source)
	if source == "" {
		source = ModelSourceManual
	}
	var id int64
	err := dao.Model.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		var insertErr error
		id, insertErr = dao.Model.Ctx(ctx).Data(do.Model{
			ProviderId: in.ProviderId,
			EndpointId: in.EndpointId,
			ModelName:  strings.TrimSpace(in.ModelName),
			Protocol:   protocol,
			Source:     source,
			Enabled:    normalizeEnabled(in.Enabled),
		}).InsertAndGetId()
		if insertErr != nil {
			return insertErr
		}
		return s.upsertModelCapability(ctx, &entity.Model{
			Id:         id,
			ProviderId: in.ProviderId,
			EndpointId: in.EndpointId,
			Protocol:   protocol,
		}, modelCapabilityFromModelSaveInput(in))
	})
	if err != nil {
		return 0, err
	}
	if err := s.InvalidateTierCache(ctx, "", "", ""); err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateModel updates one model.
func (s *serviceImpl) UpdateModel(ctx context.Context, in ModelSaveInput) error {
	if err := s.ensurePlatform(ctx); err != nil {
		return err
	}
	row, err := s.getModel(ctx, in.Id)
	if err != nil {
		return err
	}
	protocol := normalizeProtocol(in.Protocol)
	if protocol == "" || strings.TrimSpace(in.ModelName) == "" {
		return bizerr.NewCode(CodeRequestInvalid)
	}
	if _, err := s.requireProviderEndpoint(ctx, row.ProviderId, in.EndpointId, protocol); err != nil {
		return err
	}
	if err := s.ensureModelCapabilityEndpointsMatchProtocol(ctx, row.Id, row.ProviderId, protocol); err != nil {
		return err
	}
	_, err = dao.Model.Ctx(ctx).
		Where(do.Model{Id: in.Id}).
		Data(do.Model{
			ProviderId: row.ProviderId,
			EndpointId: in.EndpointId,
			ModelName:  strings.TrimSpace(in.ModelName),
			Protocol:   protocol,
			Source:     row.Source,
			Enabled:    normalizeEnabled(in.Enabled),
		}).
		Update()
	if err != nil {
		return err
	}
	return s.InvalidateTierCache(ctx, "", "", "")
}

// DeleteModel soft-deletes one model after reference checks.
func (s *serviceImpl) DeleteModel(ctx context.Context, id int64) error {
	if err := s.ensurePlatform(ctx); err != nil {
		return err
	}
	if _, err := s.getModel(ctx, id); err != nil {
		return err
	}
	inUse, err := s.modelReferenced(ctx, id)
	if err != nil {
		return err
	}
	if inUse {
		return bizerr.NewCode(CodeModelInUse)
	}
	err = dao.Model.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if _, err := dao.ModelCapability.Ctx(ctx).Where(do.ModelCapability{ModelId: id}).Delete(); err != nil {
			return err
		}
		_, err := dao.Model.Ctx(ctx).Where(do.Model{Id: id}).Delete()
		return err
	})
	if err != nil {
		return err
	}
	return s.InvalidateTierCache(ctx, "", "", "")
}

// SyncModels imports public model metadata from a provider endpoint.
func (s *serviceImpl) SyncModels(ctx context.Context, in ModelSyncInput) (*ModelSyncOutput, error) {
	if err := s.ensurePlatform(ctx); err != nil {
		return nil, err
	}
	provider, err := s.getProvider(ctx, in.ProviderId)
	if err != nil {
		return nil, err
	}
	protocol := normalizeProtocol(in.Protocol)
	if protocol == "" {
		return nil, bizerr.NewCode(CodeRequestInvalid)
	}
	endpoint, err := s.enabledEndpointForProtocol(ctx, provider.Id, protocol)
	if err != nil {
		return nil, err
	}
	names, err := s.listRemoteModels(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	out := &ModelSyncOutput{}
	existingNames, err := s.existingModelNames(ctx, provider.Id, protocol)
	if err != nil {
		return nil, err
	}
	err = dao.Model.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		for _, name := range names {
			modelName := strings.TrimSpace(name)
			if modelName == "" {
				continue
			}
			if _, ok := existingNames[modelName]; ok {
				out.Kept++
				continue
			}
			modelID, err := dao.Model.Ctx(ctx).Data(do.Model{
				ProviderId: provider.Id,
				EndpointId: endpoint.Id,
				ModelName:  modelName,
				Protocol:   protocol,
				Source:     ModelSourceAPI,
				Enabled:    enabledYes,
			}).InsertAndGetId()
			if err != nil {
				return err
			}
			if err = s.upsertModelCapability(ctx, &entity.Model{
				Id:         modelID,
				ProviderId: provider.Id,
				EndpointId: endpoint.Id,
				Protocol:   protocol,
			}, ModelCapabilitySaveInput{
				EndpointId:       endpoint.Id,
				CapabilityType:   CapabilityTypeText,
				CapabilityMethod: CapabilityMethodGenerate,
				InputModalities:  []string{"text"},
				OutputModalities: []string{"text"},
				Enabled:          enabledYes,
			}); err != nil {
				return err
			}
			existingNames[modelName] = struct{}{}
			out.Created++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err = s.InvalidateTierCache(ctx, "", "", ""); err != nil {
		return nil, err
	}
	return out, nil
}

// getProvider returns one active provider entity.
func (s *serviceImpl) getProvider(ctx context.Context, id int64) (*entity.Provider, error) {
	if id <= 0 {
		return nil, bizerr.NewCode(CodeProviderNotFound)
	}
	var row *entity.Provider
	if err := dao.Provider.Ctx(ctx).Where(do.Provider{Id: id}).Scan(&row); err != nil {
		return nil, err
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeProviderNotFound)
	}
	return row, nil
}

// getModel returns one active model entity.
func (s *serviceImpl) getModel(ctx context.Context, id int64) (*entity.Model, error) {
	if id <= 0 {
		return nil, bizerr.NewCode(CodeModelNotFound)
	}
	var row *entity.Model
	if err := dao.Model.Ctx(ctx).Where(do.Model{Id: id}).Scan(&row); err != nil {
		return nil, err
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeModelNotFound)
	}
	return row, nil
}

// collectProviderIDs extracts provider IDs from one page.
func collectProviderIDs(rows []*entity.Provider) []int64 {
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row != nil && row.Id > 0 {
			ids = append(ids, row.Id)
		}
	}
	return ids
}

// collectModelIDs extracts model IDs from one provider-scoped query result.
func collectModelIDs(rows []*entity.Model) []int64 {
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row != nil && row.Id > 0 {
			ids = append(ids, row.Id)
		}
	}
	return ids
}

// modelCapabilityFromModelSaveInput builds the default explicit capability declaration for one model row.
func modelCapabilityFromModelSaveInput(in ModelSaveInput) ModelCapabilitySaveInput {
	capabilityType := normalizeCapabilityType(in.CapabilityType)
	capabilityMethod := normalizeCapabilityMethod(in.CapabilityMethod)
	inputModalities, outputModalities := defaultModalitiesForCapability(capabilityType, capabilityMethod)
	supportsOperation := enabledNo
	if strings.HasPrefix(capabilityMethod, "operation.") {
		supportsOperation = enabledYes
	}
	return ModelCapabilitySaveInput{
		EndpointId:        in.EndpointId,
		CapabilityType:    capabilityType,
		CapabilityMethod:  capabilityMethod,
		InputModalities:   inputModalities,
		OutputModalities:  outputModalities,
		MaxInputTokens:    in.MaxInputTokens,
		MaxOutputTokens:   in.MaxOutputTokens,
		SupportsOperation: supportsOperation,
		SupportsThinking:  in.SupportsThinking,
		SupportedEfforts:  in.SupportedEfforts,
		Enabled:           normalizeEnabled(in.Enabled),
	}
}

// defaultModalitiesForCapability returns conservative modality projections for a capability method.
func defaultModalitiesForCapability(capabilityType string, capabilityMethod string) ([]string, []string) {
	switch capabilityType {
	case "audio":
		if capabilityMethod == "transcribe" {
			return []string{"audio"}, []string{"text"}
		}
		return []string{"text"}, []string{"audio"}
	case "document":
		return []string{"document"}, []string{"text"}
	case "embedding":
		return []string{"text"}, []string{"embedding"}
	case "image":
		if capabilityMethod == "edit" {
			return []string{"image", "text"}, []string{"image"}
		}
		return []string{"text"}, []string{"image"}
	case "safety":
		return []string{"text"}, []string{"safety"}
	case "video":
		return []string{"text"}, []string{"video"}
	case "vision":
		return []string{"image"}, []string{"text"}
	default:
		return []string{"text"}, []string{"text"}
	}
}

// countModelsByProvider counts all and enabled models using one batch query.
func (s *serviceImpl) countModelsByProvider(ctx context.Context, providerIDs []int64) (map[int64]int, map[int64]int, error) {
	counts := make(map[int64]int, len(providerIDs))
	enabledCounts := make(map[int64]int, len(providerIDs))
	if len(providerIDs) == 0 {
		return counts, enabledCounts, nil
	}
	cols := dao.Model.Columns()
	type modelCountRow struct {
		ProviderId        int64 `orm:"provider_id"`
		ModelCount        int64 `orm:"model_count"`
		EnabledModelCount int64 `orm:"enabled_model_count"`
	}
	rows := make([]modelCountRow, 0)
	if err := dao.Model.Ctx(ctx).
		Fields(cols.ProviderId, "COUNT(*) AS model_count", "SUM(CASE WHEN "+cols.Enabled+" = 1 THEN 1 ELSE 0 END) AS enabled_model_count").
		WhereIn(cols.ProviderId, providerIDs).
		Group(cols.ProviderId).
		Scan(&rows); err != nil {
		return nil, nil, err
	}
	for _, row := range rows {
		counts[row.ProviderId] = int(row.ModelCount)
		enabledCounts[row.ProviderId] = int(row.EnabledModelCount)
	}
	return counts, enabledCounts, nil
}

// listModelSummariesByProvider returns compact model summaries for the current provider page.
func (s *serviceImpl) listModelSummariesByProvider(ctx context.Context, providerIDs []int64) (map[int64][]*ProviderModelSummaryItem, error) {
	result := make(map[int64][]*ProviderModelSummaryItem, len(providerIDs))
	if len(providerIDs) == 0 {
		return result, nil
	}
	cols := dao.Model.Columns()
	rows := make([]*entity.Model, 0)
	if err := dao.Model.Ctx(ctx).
		Fields(cols.Id, cols.ProviderId, cols.ModelName, cols.Protocol, cols.Enabled).
		WhereIn(cols.ProviderId, providerIDs).
		OrderAsc(cols.ProviderId).
		OrderAsc(cols.Id).
		Scan(&rows); err != nil {
		return nil, err
	}
	modelsByID := make(map[int64]*entity.Model, len(rows))
	modelIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		modelsByID[row.Id] = row
		modelIDs = append(modelIDs, row.Id)
	}
	capabilities, err := s.modelCapabilitiesByModel(ctx, modelIDs)
	if err != nil {
		return nil, err
	}
	for _, capability := range capabilities {
		model := modelsByID[capability.ModelId]
		summary := providerModelSummaryFromRows(model, capability)
		if summary == nil {
			continue
		}
		result[model.ProviderId] = append(result[model.ProviderId], summary)
	}
	return result, nil
}

// modelCapabilitiesByModelMethod loads one method capability per model in a single query.
func (s *serviceImpl) modelCapabilitiesByModelMethod(
	ctx context.Context,
	modelIDs []int64,
	capabilityType string,
	capabilityMethod string,
) (map[int64]*entity.ModelCapability, error) {
	result := make(map[int64]*entity.ModelCapability, len(modelIDs))
	if len(modelIDs) == 0 {
		return result, nil
	}
	rows := make([]*entity.ModelCapability, 0)
	if err := dao.ModelCapability.Ctx(ctx).
		WhereIn(dao.ModelCapability.Columns().ModelId, modelIDs).
		Where(do.ModelCapability{
			CapabilityType:   normalizeCapabilityType(capabilityType),
			CapabilityMethod: normalizeCapabilityMethod(capabilityMethod),
		}).
		Scan(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row != nil {
			result[row.ModelId] = row
		}
	}
	return result, nil
}

// modelCapabilitiesByModel loads all method capability rows for a model set in one query.
func (s *serviceImpl) modelCapabilitiesByModel(ctx context.Context, modelIDs []int64) ([]*entity.ModelCapability, error) {
	if len(modelIDs) == 0 {
		return nil, nil
	}
	cols := dao.ModelCapability.Columns()
	rows := make([]*entity.ModelCapability, 0)
	if err := dao.ModelCapability.Ctx(ctx).
		WhereIn(cols.ModelId, modelIDs).
		OrderAsc(cols.ModelId).
		OrderAsc(cols.CapabilityType).
		OrderAsc(cols.CapabilityMethod).
		Scan(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// ensureModelCapabilityEndpointsMatchProtocol rejects model protocol changes that would leave method endpoint overrides inconsistent.
func (s *serviceImpl) ensureModelCapabilityEndpointsMatchProtocol(ctx context.Context, modelID int64, providerID int64, protocol string) error {
	capRows := make([]*entity.ModelCapability, 0)
	capCols := dao.ModelCapability.Columns()
	if err := dao.ModelCapability.Ctx(ctx).
		Fields(capCols.EndpointId).
		Where(do.ModelCapability{ModelId: modelID}).
		WhereGT(capCols.EndpointId, 0).
		Scan(&capRows); err != nil {
		return err
	}
	endpointIDs := make([]int64, 0, len(capRows))
	for _, row := range capRows {
		if row != nil && row.EndpointId > 0 {
			endpointIDs = append(endpointIDs, row.EndpointId)
		}
	}
	endpoints, err := s.endpointsByID(ctx, endpointIDs)
	if err != nil {
		return err
	}
	for _, row := range capRows {
		endpoint := endpoints[row.EndpointId]
		if endpoint == nil || endpoint.ProviderId != providerID || endpoint.Protocol != protocol {
			return bizerr.NewCode(CodeProviderProtocolRequired)
		}
	}
	return nil
}

// providerReferenced reports whether active tier bindings reference one provider.
func (s *serviceImpl) providerReferenced(ctx context.Context, providerID int64) (bool, error) {
	count, err := dao.TierBinding.Ctx(ctx).Where(do.TierBinding{ProviderId: providerID}).Count()
	return count > 0, err
}

// modelReferenced reports whether active tier bindings reference one model.
func (s *serviceImpl) modelReferenced(ctx context.Context, modelID int64) (bool, error) {
	count, err := dao.TierBinding.Ctx(ctx).Where(do.TierBinding{ModelId: modelID}).Count()
	return count > 0, err
}

// existingModelNames returns provider model names for one protocol in a single query.
func (s *serviceImpl) existingModelNames(ctx context.Context, providerID int64, protocol string) (map[string]struct{}, error) {
	result := make(map[string]struct{})
	rows := make([]*entity.Model, 0)
	cols := dao.Model.Columns()
	if err := dao.Model.Ctx(ctx).
		Fields(cols.ModelName).
		Where(do.Model{
			ProviderId: providerID,
			Protocol:   protocol,
		}).
		Scan(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		name := strings.TrimSpace(row.ModelName)
		if name != "" {
			result[name] = struct{}{}
		}
	}
	return result, nil
}
