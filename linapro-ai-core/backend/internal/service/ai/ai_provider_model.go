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

// CreateProvider creates one provider and optional provider-form endpoints.
func (s *serviceImpl) CreateProvider(ctx context.Context, in ProviderSaveInput) (int64, error) {
	if err := s.ensurePlatform(ctx); err != nil {
		return 0, err
	}
	if strings.TrimSpace(in.Name) == "" {
		return 0, bizerr.NewCode(CodeRequestInvalid)
	}
	var id int64
	err := dao.Provider.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		var insertErr error
		id, insertErr = dao.Provider.Ctx(ctx).Data(do.Provider{
			Name:       strings.TrimSpace(in.Name),
			WebsiteUrl: strings.TrimSpace(in.WebsiteUrl),
			Remark:     strings.TrimSpace(in.Remark),
			Enabled:    normalizeEnabled(in.Enabled),
		}).InsertAndGetId()
		if insertErr != nil {
			return insertErr
		}
		return s.syncProviderFormEndpoints(ctx, id, in.Endpoints)
	})
	if err != nil {
		return 0, err
	}
	if err = s.invalidateTierCache(ctx, "", "", ""); err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateProvider updates one provider and optional provider-form endpoints.
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
	err = dao.Provider.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if _, err = dao.Provider.Ctx(ctx).
			Where(do.Provider{Id: in.Id}).
			Data(do.Provider{
				Name:       strings.TrimSpace(in.Name),
				WebsiteUrl: strings.TrimSpace(in.WebsiteUrl),
				Remark:     strings.TrimSpace(in.Remark),
				Enabled:    normalizeEnabled(in.Enabled),
			}).
			Update(); err != nil {
			return err
		}
		return s.syncProviderFormEndpoints(ctx, in.Id, in.Endpoints)
	})
	if err != nil {
		return err
	}
	return s.invalidateTierCache(ctx, "", "", "")
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
	return s.invalidateTierCache(ctx, "", "", "")
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
	if in.Enabled != nil {
		enabled := normalizeEnabled(*in.Enabled)
		model = model.Where(cols.Enabled, enabled)
	}
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	rows := make([]*entity.Model, 0)
	if err = model.Page(pageNum, pageSize).OrderAsc(cols.Id).Scan(&rows); err != nil {
		return nil, err
	}
	items := make([]*ModelItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, modelToItem(row, nil))
	}
	return &ModelListOutput{List: items, Total: total}, nil
}

// ListAllModels returns one bounded model-dimension page with provider and endpoint projections.
func (s *serviceImpl) ListAllModels(ctx context.Context, in ModelGlobalListInput) (*ModelListOutput, error) {
	if err := s.ensurePlatform(ctx); err != nil {
		return nil, err
	}
	pageNum, pageSize := normalizePage(in.PageNum, in.PageSize)
	cols := dao.Model.Columns()
	model := dao.Model.Ctx(ctx)
	if in.ProviderId > 0 {
		model = model.Where(cols.ProviderId, in.ProviderId)
	}
	if keyword := strings.TrimSpace(in.Keyword); keyword != "" {
		model = model.WhereLike(cols.ModelName, "%"+keyword+"%")
	}
	if in.Enabled != nil {
		enabled := normalizeEnabled(*in.Enabled)
		model = model.Where(cols.Enabled, enabled)
	}
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	rows := make([]*entity.Model, 0)
	if err = model.Page(pageNum, pageSize).OrderDesc(cols.Id).Scan(&rows); err != nil {
		return nil, err
	}
	providers, err := s.providersByID(ctx, collectModelProviderIDs(rows))
	if err != nil {
		return nil, err
	}
	endpoints, err := s.endpointsByID(ctx, collectModelEndpointIDs(rows))
	if err != nil {
		return nil, err
	}
	items := make([]*ModelItem, 0, len(rows))
	for _, row := range rows {
		item := modelToItem(row, nil)
		if item == nil {
			continue
		}
		if provider := providers[row.ProviderId]; provider != nil {
			item.ProviderName = provider.Name
		}
		if endpoint := endpoints[item.EndpointId]; endpoint != nil {
			item.EndpointBaseUrl = endpoint.BaseUrl
		}
		items = append(items, item)
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
		return nil
	})
	if err != nil {
		return 0, err
	}
	if err := s.invalidateTierCache(ctx, "", "", ""); err != nil {
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
	return s.invalidateTierCache(ctx, "", "", "")
}

// DeleteModel soft-deletes all provider-local rows sharing the target model name
// after reference checks.
func (s *serviceImpl) DeleteModel(ctx context.Context, id int64) error {
	if err := s.ensurePlatform(ctx); err != nil {
		return err
	}
	row, err := s.getModel(ctx, id)
	if err != nil {
		return err
	}
	modelRows, err := s.modelsByProviderAndName(ctx, row.ProviderId, row.ModelName)
	if err != nil {
		return err
	}
	modelIDs := collectModelIDs(modelRows)
	if len(modelIDs) == 0 {
		return bizerr.NewCode(CodeModelNotFound)
	}
	inUse, err := s.modelsReferenced(ctx, modelIDs)
	if err != nil {
		return err
	}
	if inUse {
		return bizerr.NewCode(CodeModelInUse)
	}
	err = dao.Model.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if _, err := dao.ModelCapability.Ctx(ctx).
			WhereIn(dao.ModelCapability.Columns().ModelId, modelIDs).
			Delete(); err != nil {
			return err
		}
		_, err := dao.Model.Ctx(ctx).
			WhereIn(dao.Model.Columns().Id, modelIDs).
			Delete()
		return err
	})
	if err != nil {
		return err
	}
	return s.invalidateTierCache(ctx, "", "", "")
}

// SyncModels imports public model metadata from provider endpoints.
func (s *serviceImpl) SyncModels(ctx context.Context, in ModelSyncInput) (*ModelSyncOutput, error) {
	if err := s.ensurePlatform(ctx); err != nil {
		return nil, err
	}
	provider, err := s.getProvider(ctx, in.ProviderId)
	if err != nil {
		return nil, err
	}
	protocol := normalizeProtocol(in.Protocol)
	if strings.TrimSpace(in.Protocol) != "" && protocol == "" {
		return nil, bizerr.NewCode(CodeRequestInvalid)
	}
	endpoints, err := s.enabledSyncEndpoints(ctx, provider.Id, protocol)
	if err != nil {
		return nil, err
	}
	out := &ModelSyncOutput{}
	successes := make([]modelSyncEndpointResult, 0, len(endpoints))
	var lastErr error
	for _, endpoint := range endpoints {
		models, listErr := s.listRemoteModels(ctx, endpoint)
		if listErr != nil {
			lastErr = listErr
			continue
		}
		successes = append(successes, modelSyncEndpointResult{endpoint: endpoint, models: models})
	}
	if len(successes) == 0 {
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, bizerr.NewCode(CodeProviderProtocolRequired)
	}
	existingModelsByProtocol, err := s.existingModelsByProtocol(ctx, provider.Id, collectSyncProtocols(successes))
	if err != nil {
		return nil, err
	}
	err = dao.Model.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		for _, success := range successes {
			existingModels := existingModelsByProtocol[success.endpoint.Protocol]
			if existingModels == nil {
				existingModels = make(map[string]*entity.Model)
				existingModelsByProtocol[success.endpoint.Protocol] = existingModels
			}
			for _, remote := range success.models {
				modelName := strings.TrimSpace(remote.Name)
				if modelName == "" {
					continue
				}
				if existing := existingModels[modelName]; existing != nil {
					out.Kept++
					continue
				}
				modelID, err := dao.Model.Ctx(ctx).Data(do.Model{
					ProviderId: provider.Id,
					EndpointId: success.endpoint.Id,
					ModelName:  modelName,
					Protocol:   success.endpoint.Protocol,
					Source:     modelSourceAPI,
					Enabled:    enabledYes,
				}).InsertAndGetId()
				if err != nil {
					return err
				}
				existingModels[modelName] = &entity.Model{
					Id:         modelID,
					ProviderId: provider.Id,
					EndpointId: success.endpoint.Id,
					Protocol:   success.endpoint.Protocol,
				}
				out.Created++
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err = s.invalidateTierCache(ctx, "", "", ""); err != nil {
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

// modelsByProviderAndName returns active model rows in one provider sharing the
// same display model name.
func (s *serviceImpl) modelsByProviderAndName(ctx context.Context, providerID int64, modelName string) ([]*entity.Model, error) {
	name := strings.TrimSpace(modelName)
	if providerID <= 0 || name == "" {
		return nil, nil
	}
	cols := dao.Model.Columns()
	rows := make([]*entity.Model, 0)
	if err := dao.Model.Ctx(ctx).
		Fields(cols.Id, cols.ProviderId, cols.ModelName).
		Where(do.Model{ProviderId: providerID, ModelName: name}).
		OrderAsc(cols.Id).
		Scan(&rows); err != nil {
		return nil, err
	}
	return rows, nil
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

// collectModelProviderIDs extracts unique provider IDs from one model page.
func collectModelProviderIDs(rows []*entity.Model) []int64 {
	seen := make(map[int64]struct{}, len(rows))
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.ProviderId <= 0 {
			continue
		}
		if _, ok := seen[row.ProviderId]; ok {
			continue
		}
		seen[row.ProviderId] = struct{}{}
		ids = append(ids, row.ProviderId)
	}
	return ids
}

// collectModelEndpointIDs extracts endpoint IDs needed by one model page projection.
func collectModelEndpointIDs(rows []*entity.Model) []int64 {
	seen := make(map[int64]struct{}, len(rows))
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		endpointID := row.EndpointId
		if endpointID <= 0 {
			continue
		}
		if _, ok := seen[endpointID]; ok {
			continue
		}
		seen[endpointID] = struct{}{}
		ids = append(ids, endpointID)
	}
	return ids
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
	summariesByProviderName := make(map[int64]map[string]*ProviderModelSummaryItem, len(providerIDs))
	for _, model := range rows {
		if model == nil {
			continue
		}
		modelName := strings.TrimSpace(model.ModelName)
		if modelName == "" {
			continue
		}
		providerSummaries := summariesByProviderName[model.ProviderId]
		if providerSummaries == nil {
			providerSummaries = make(map[string]*ProviderModelSummaryItem)
			summariesByProviderName[model.ProviderId] = providerSummaries
		}
		if summary := providerSummaries[modelName]; summary != nil {
			if summary.Enabled != enabledYes && model.Enabled == enabledYes {
				summary.Enabled = enabledYes
			}
			continue
		}
		summary := providerModelSummaryFromRows(model, nil)
		if summary == nil {
			continue
		}
		summary.ModelName = modelName
		providerSummaries[modelName] = summary
		result[model.ProviderId] = append(result[model.ProviderId], summary)
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

// providerReferenced reports whether active tier bindings reference one provider.
func (s *serviceImpl) providerReferenced(ctx context.Context, providerID int64) (bool, error) {
	count, err := dao.TierBinding.Ctx(ctx).Where(do.TierBinding{ProviderId: providerID}).Count()
	return count > 0, err
}

// modelsReferenced reports whether active tier bindings reference any model in the set.
func (s *serviceImpl) modelsReferenced(ctx context.Context, modelIDs []int64) (bool, error) {
	if len(modelIDs) == 0 {
		return false, nil
	}
	count, err := dao.TierBinding.Ctx(ctx).
		WhereIn(dao.TierBinding.Columns().ModelId, modelIDs).
		Count()
	return count > 0, err
}

// existingModelsByProtocol returns provider models grouped by protocol and model name in one query.
func (s *serviceImpl) existingModelsByProtocol(ctx context.Context, providerID int64, protocols []string) (map[string]map[string]*entity.Model, error) {
	result := make(map[string]map[string]*entity.Model, len(protocols))
	if providerID <= 0 || len(protocols) == 0 {
		return result, nil
	}
	rows := make([]*entity.Model, 0)
	cols := dao.Model.Columns()
	if err := dao.Model.Ctx(ctx).
		Fields(cols.Id, cols.ProviderId, cols.EndpointId, cols.ModelName, cols.Protocol, cols.Enabled).
		Where(cols.ProviderId, providerID).
		WhereIn(cols.Protocol, protocols).
		Scan(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		protocol := normalizeProtocol(row.Protocol)
		modelName := strings.TrimSpace(row.ModelName)
		if protocol == "" || modelName == "" {
			continue
		}
		models := result[protocol]
		if models == nil {
			models = make(map[string]*entity.Model)
			result[protocol] = models
		}
		models[modelName] = row
	}
	return result, nil
}

type modelSyncEndpointResult struct {
	endpoint *entity.ProviderEndpoint
	models   []remoteModel
}

func collectSyncProtocols(items []modelSyncEndpointResult) []string {
	seen := make(map[string]struct{}, len(items))
	protocols := make([]string, 0, len(items))
	for _, item := range items {
		if item.endpoint == nil {
			continue
		}
		protocol := normalizeProtocol(item.endpoint.Protocol)
		if protocol == "" {
			continue
		}
		if _, ok := seen[protocol]; ok {
			continue
		}
		seen[protocol] = struct{}{}
		protocols = append(protocols, protocol)
	}
	return protocols
}
