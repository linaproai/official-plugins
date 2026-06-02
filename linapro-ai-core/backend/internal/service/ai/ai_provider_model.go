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

// ListProviders returns a platform-only paged provider list with model counts.
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
	counts, enabledCounts, err := s.countModelsByProvider(ctx, collectProviderIDs(rows))
	if err != nil {
		return nil, err
	}
	items := make([]*ProviderItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, providerToItem(row, counts[row.Id], enabledCounts[row.Id]))
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
	return providerToItem(row, counts[id], enabledCounts[id]), nil
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
		Name:             strings.TrimSpace(in.Name),
		WebsiteUrl:       strings.TrimSpace(in.WebsiteUrl),
		Remark:           strings.TrimSpace(in.Remark),
		OpenaiBaseUrl:    strings.TrimSpace(in.OpenaiBaseUrl),
		AnthropicBaseUrl: strings.TrimSpace(in.AnthropicBaseUrl),
		ApiKeySecretRef:  strings.TrimSpace(in.ApiKeySecretRef),
		Enabled:          normalizeEnabled(in.Enabled),
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}
	if err = s.InvalidateTierCache(ctx, "", ""); err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateProvider updates one provider.
func (s *serviceImpl) UpdateProvider(ctx context.Context, in ProviderSaveInput) error {
	if err := s.ensurePlatform(ctx); err != nil {
		return err
	}
	row, err := s.getProvider(ctx, in.Id)
	if err != nil {
		return err
	}
	if strings.TrimSpace(in.Name) == "" {
		return bizerr.NewCode(CodeRequestInvalid)
	}
	secretRef := strings.TrimSpace(in.ApiKeySecretRef)
	if shouldKeepExistingSecret(secretRef) {
		secretRef = row.ApiKeySecretRef
	}
	_, err = dao.Provider.Ctx(ctx).
		Where(do.Provider{Id: in.Id}).
		Data(do.Provider{
			Name:             strings.TrimSpace(in.Name),
			WebsiteUrl:       strings.TrimSpace(in.WebsiteUrl),
			Remark:           strings.TrimSpace(in.Remark),
			OpenaiBaseUrl:    strings.TrimSpace(in.OpenaiBaseUrl),
			AnthropicBaseUrl: strings.TrimSpace(in.AnthropicBaseUrl),
			ApiKeySecretRef:  secretRef,
			Enabled:          normalizeEnabled(in.Enabled),
		}).
		Update()
	if err != nil {
		return err
	}
	return s.InvalidateTierCache(ctx, "", "")
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
		if _, err := dao.Model.Ctx(ctx).Where(do.Model{ProviderId: id}).Delete(); err != nil {
			return err
		}
		_, err := dao.Provider.Ctx(ctx).Where(do.Provider{Id: id}).Delete()
		return err
	})
	if err != nil {
		return err
	}
	return s.InvalidateTierCache(ctx, "", "")
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
	if capabilityType := normalizeCapabilityType(in.CapabilityType); capabilityType != "" {
		model = model.Where(cols.CapabilityType, capabilityType)
	}
	if in.Enabled != nil {
		model = model.Where(cols.Enabled, normalizeEnabled(*in.Enabled))
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
		items = append(items, modelToItem(row))
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
	efforts, _, err := normalizeEfforts(in.SupportedEfforts)
	if err != nil {
		return 0, err
	}
	protocol := normalizeProtocol(in.Protocol)
	if protocol == "" || strings.TrimSpace(in.ModelName) == "" {
		return 0, bizerr.NewCode(CodeRequestInvalid)
	}
	source := strings.TrimSpace(in.Source)
	if source == "" {
		source = ModelSourceManual
	}
	id, err := dao.Model.Ctx(ctx).Data(do.Model{
		ProviderId:       in.ProviderId,
		CapabilityType:   normalizeCapabilityType(in.CapabilityType),
		ModelName:        strings.TrimSpace(in.ModelName),
		Protocol:         protocol,
		Source:           source,
		SupportsThinking: normalizeEnabled(in.SupportsThinking),
		SupportedEfforts: efforts,
		MaxInputTokens:   in.MaxInputTokens,
		MaxOutputTokens:  in.MaxOutputTokens,
		Enabled:          normalizeEnabled(in.Enabled),
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}
	if err = s.InvalidateTierCache(ctx, "", ""); err != nil {
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
	efforts, _, err := normalizeEfforts(in.SupportedEfforts)
	if err != nil {
		return err
	}
	protocol := normalizeProtocol(in.Protocol)
	if protocol == "" || strings.TrimSpace(in.ModelName) == "" {
		return bizerr.NewCode(CodeRequestInvalid)
	}
	_, err = dao.Model.Ctx(ctx).
		Where(do.Model{Id: in.Id}).
		Data(do.Model{
			ProviderId:       row.ProviderId,
			CapabilityType:   normalizeCapabilityType(in.CapabilityType),
			ModelName:        strings.TrimSpace(in.ModelName),
			Protocol:         protocol,
			Source:           row.Source,
			SupportsThinking: normalizeEnabled(in.SupportsThinking),
			SupportedEfforts: efforts,
			MaxInputTokens:   in.MaxInputTokens,
			MaxOutputTokens:  in.MaxOutputTokens,
			Enabled:          normalizeEnabled(in.Enabled),
		}).
		Update()
	if err != nil {
		return err
	}
	return s.InvalidateTierCache(ctx, "", "")
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
	if _, err = dao.Model.Ctx(ctx).Where(do.Model{Id: id}).Delete(); err != nil {
		return err
	}
	return s.InvalidateTierCache(ctx, "", "")
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
	names, err := s.listRemoteModels(ctx, provider, protocol)
	if err != nil {
		return nil, err
	}
	out := &ModelSyncOutput{}
	existingNames, err := s.existingModelNames(ctx, provider.Id, protocol)
	if err != nil {
		return nil, err
	}
	for _, name := range names {
		modelName := strings.TrimSpace(name)
		if modelName == "" {
			continue
		}
		if _, ok := existingNames[modelName]; ok {
			out.Kept++
			continue
		}
		if _, err = dao.Model.Ctx(ctx).Data(do.Model{
			ProviderId:       provider.Id,
			CapabilityType:   CapabilityTypeText,
			ModelName:        modelName,
			Protocol:         protocol,
			Source:           ModelSourceAPI,
			SupportsThinking: enabledNo,
			SupportedEfforts: "",
			Enabled:          enabledYes,
		}).InsertAndGetId(); err != nil {
			return nil, err
		}
		existingNames[modelName] = struct{}{}
		out.Created++
	}
	if err = s.InvalidateTierCache(ctx, "", ""); err != nil {
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
		Where(do.Model{ProviderId: providerID, Protocol: protocol}).
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
