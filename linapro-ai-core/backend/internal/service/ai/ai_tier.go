// This file implements tier management and the short-lived tier binding cache.

package ai

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/aicap/aitext"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/cachecap"
	"lina-plugin-linapro-ai-core/backend/internal/dao"
	"lina-plugin-linapro-ai-core/backend/internal/model/do"
	"lina-plugin-linapro-ai-core/backend/internal/model/entity"
)

// resolvedTierBinding contains the public and secret metadata needed for one provider call.
type resolvedTierBinding struct {
	TierId            int64
	TierCode          string
	CapabilityType    string
	CapabilityMethod  string
	DefaultEffort     string
	ProviderId        int64
	ProviderName      string
	ModelId           int64
	ModelName         string
	Protocol          string
	EndpointId        int64
	EndpointBaseUrl   string
	EndpointSecretRef string
}

// ListTiers returns the fixed AI tier list for one capability method with binding projections.
func (s *serviceImpl) ListTiers(ctx context.Context, capabilityType string, capabilityMethod string) ([]*TierItem, error) {
	if err := s.ensurePlatform(ctx); err != nil {
		return nil, err
	}
	capabilityType = normalizeCapabilityType(capabilityType)
	capabilityMethod = normalizeCapabilityMethod(capabilityMethod)
	rows := make([]*entity.Tier, 0)
	if err := dao.Tier.Ctx(ctx).
		Where(do.Tier{CapabilityType: capabilityType, CapabilityMethod: capabilityMethod}).
		OrderAsc(dao.Tier.Columns().SortOrder).
		Scan(&rows); err != nil {
		return nil, err
	}
	bindings, err := s.primaryBindingsByTier(ctx, collectTierIDs(rows))
	if err != nil {
		return nil, err
	}
	items := make([]*TierItem, 0, len(rows))
	for _, row := range rows {
		item := &TierItem{
			Id:                   row.Id,
			CapabilityType:       row.CapabilityType,
			CapabilityMethod:     row.CapabilityMethod,
			Code:                 row.Code,
			DisplayName:          row.DisplayName,
			Description:          row.Description,
			DefaultEffort:        row.DefaultEffort,
			Enabled:              row.Enabled,
			SortOrder:            row.SortOrder,
			LastTestStatus:       row.LastTestStatus,
			LastTestLatencyMs:    row.LastTestLatencyMs,
			LastTestErrorSummary: row.LastTestErrorSummary,
			LastTestAt:           row.LastTestAt,
			UpdatedAt:            row.UpdatedAt,
		}
		if binding := bindings[row.Id]; binding != nil {
			item.Binding = &TierBindingItem{
				ProviderId:   binding.ProviderId,
				ProviderName: binding.ProviderName,
				ModelId:      binding.ModelId,
				ModelName:    binding.ModelName,
				Protocol:     binding.Protocol,
				Enabled:      enabledYes,
			}
		}
		items = append(items, item)
	}
	return items, nil
}

// UpdateTier updates one fixed tier and its primary binding.
func (s *serviceImpl) UpdateTier(ctx context.Context, in TierUpdateInput) error {
	if err := s.ensurePlatform(ctx); err != nil {
		return err
	}
	capabilityType := normalizeCapabilityType(in.CapabilityType)
	capabilityMethod := normalizeCapabilityMethod(in.CapabilityMethod)
	code := normalizeTierCode(in.Code)
	if code == "" {
		return bizerr.NewCode(CodeTierNotFound)
	}
	effort, err := normalizeEffort(in.DefaultEffort)
	if err != nil {
		return err
	}
	tier, err := s.getTier(ctx, capabilityType, capabilityMethod, code)
	if err != nil {
		return err
	}
	var model *entity.Model
	bindingRequested := in.ProviderId > 0 || in.ModelId > 0
	if bindingRequested && (in.ProviderId <= 0 || in.ModelId <= 0) {
		return bizerr.NewCode(CodeRequestInvalid)
	}
	if bindingRequested {
		model, _, err = s.validateModelBinding(ctx, in.ProviderId, in.ModelId)
		if err != nil {
			return err
		}
		if err = s.validateModelCapabilityBinding(ctx, model.Id, capabilityType, capabilityMethod); err != nil {
			return err
		}
	} else if normalizeEnabled(in.Enabled) == enabledYes {
		bindings, err := s.primaryBindingsByTier(ctx, []int64{tier.Id})
		if err != nil {
			return err
		}
		if bindings[tier.Id] == nil {
			return bizerr.NewCode(CodeTierBindingUnavailable)
		}
		if !effortSupportedByBinding(bindings[tier.Id], effort) {
			return bizerr.NewCode(CodeThinkingEffortUnsupported)
		}
	}
	err = dao.Tier.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if _, err := dao.Tier.Ctx(ctx).
			Where(do.Tier{Id: tier.Id}).
			Data(do.Tier{
				DefaultEffort: effort,
				Enabled:       normalizeEnabled(in.Enabled),
			}).
			Update(); err != nil {
			return err
		}
		if model != nil {
			return s.upsertPrimaryBinding(ctx, tier.Id, in.ProviderId, in.ModelId)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return s.InvalidateTierCache(ctx, capabilityType, capabilityMethod, code)
}

// TestTier executes a saved or draft tier binding test.
func (s *serviceImpl) TestTier(ctx context.Context, in TierTestInput) (*TierTestOutput, error) {
	if err := s.ensurePlatform(ctx); err != nil {
		return nil, err
	}
	capabilityType := normalizeCapabilityType(in.CapabilityType)
	capabilityMethod := normalizeCapabilityMethod(in.CapabilityMethod)
	code := normalizeTierCode(in.Code)
	if code == "" {
		return nil, bizerr.NewCode(CodeTierNotFound)
	}
	tier, err := s.getTier(ctx, capabilityType, capabilityMethod, code)
	if err != nil {
		return nil, err
	}
	effort, err := normalizeEffort(in.ThinkingEffort)
	if err != nil {
		return nil, err
	}
	var binding *resolvedTierBinding
	draftBindingRequested := in.ProviderId > 0 || in.ModelId > 0
	if draftBindingRequested {
		model, endpoint, err := s.validateModelBinding(ctx, in.ProviderId, in.ModelId)
		if err != nil {
			return nil, err
		}
		if err = s.validateModelCapabilityBinding(ctx, model.Id, capabilityType, capabilityMethod); err != nil {
			return nil, err
		}
		provider, err := s.getProvider(ctx, in.ProviderId)
		if err != nil {
			return nil, err
		}
		binding = resolvedBindingFromRows(tier, provider, model, endpoint)
	} else {
		binding, err = s.resolveTierBinding(ctx, capabilityType, capabilityMethod, code)
		if err != nil {
			return nil, err
		}
		if effort == "" {
			effort = binding.DefaultEffort
		}
		if !effortSupportedByBinding(binding, effort) {
			return nil, bizerr.NewCode(CodeThinkingEffortUnsupported)
		}
	}
	if effort == "" {
		effort = binding.DefaultEffort
	}
	messages := in.Messages
	if len(messages) == 0 {
		messages = defaultTierTestMessages()
	}
	testedAt := time.Now()
	callCtx, cancel := context.WithTimeout(ctx, normalizeTierTestTimeout(in.timeout))
	defer cancel()
	result, callErr := s.callProvider(callCtx, binding, messages, in.MaxOutputTokens, nil, effort)
	usage := aitext.Usage{}
	output := &TierTestOutput{
		Status:         InvocationStatusSuccess,
		ProviderName:   binding.ProviderName,
		ModelName:      binding.ModelName,
		Protocol:       binding.Protocol,
		ThinkingEffort: effort,
		TestedAt:       &testedAt,
	}
	if result != nil {
		usage = result.Usage
		output.LatencyMs = result.LatencyMs
		output.ThinkingEffort = result.ThinkingEffort
	}
	if callErr != nil {
		output.Status = InvocationStatusFailed
		output.ErrorSummary = sanitizeErrorSummary(callErr)
	}
	invocationEffort := output.ThinkingEffort
	if invocationEffort == "" {
		invocationEffort = effort
	}
	s.writeInvocationRecord(ctx, invocationWriteInput{
		requestID:        requestIDFromMetadata(nil),
		capabilityType:   capabilityType,
		capabilityMethod: capabilityMethod,
		purpose:          InvocationPurposeTierTest,
		tierCode:         code,
		sourcePluginID:   aitext.ProviderPluginID,
		thinkingEffort:   invocationEffort,
		binding:          binding,
		status:           output.Status,
		usage:            usage,
		latencyMs:        output.LatencyMs,
		err:              callErr,
	})
	if !draftBindingRequested {
		_, updateErr := dao.Tier.Ctx(ctx).
			Where(do.Tier{Id: tier.Id}).
			Data(do.Tier{
				LastTestStatus:       output.Status,
				LastTestLatencyMs:    output.LatencyMs,
				LastTestErrorSummary: output.ErrorSummary,
				LastTestAt:           &testedAt,
			}).
			Update()
		if updateErr != nil {
			return nil, updateErr
		}
	}
	if callErr != nil {
		return output, nil
	}
	return output, nil
}

// normalizeTierTestTimeout returns the explicit timeout or the default test timeout.
func normalizeTierTestTimeout(timeout time.Duration) time.Duration {
	if timeout > 0 {
		return timeout
	}
	return tierTestTimeout
}

// InvalidateTierCache publishes a shared revision and clears local cached tier binding entries.
func (s *serviceImpl) InvalidateTierCache(ctx context.Context, capabilityType string, capabilityMethod string, tierCode string) error {
	if s == nil {
		return nil
	}
	revision, err := s.publishTierCacheRevision(ctx)
	if err != nil {
		return err
	}
	s.clearLocalTierCache(revision, capabilityType, capabilityMethod, tierCode)
	return nil
}

// publishTierCacheRevision increments the plugin-scoped shared tier cache revision.
func (s *serviceImpl) publishTierCacheRevision(ctx context.Context) (int64, error) {
	if s == nil || s.cacheSvc == nil {
		return 0, nil
	}
	item, err := s.cacheSvc.Incr(ctx, tierCacheNamespace, tierCacheRevisionKey, 1, 0)
	if err != nil {
		return 0, gerror.Wrap(err, "publish AI tier cache revision failed")
	}
	return cacheRevisionValue(item), nil
}

// observeTierCacheRevision clears this process cache after another instance publishes a revision.
func (s *serviceImpl) observeTierCacheRevision(ctx context.Context) error {
	if s == nil || s.cacheSvc == nil {
		return nil
	}
	item, found, err := s.cacheSvc.Get(ctx, tierCacheNamespace, tierCacheRevisionKey)
	if err != nil {
		return gerror.Wrap(err, "read AI tier cache revision failed")
	}
	revision := int64(0)
	if found {
		revision = cacheRevisionValue(item)
	}
	s.cacheMu.RLock()
	current := s.revision
	s.cacheMu.RUnlock()
	if revision == current {
		return nil
	}
	s.clearLocalTierCache(revision, "", "", "")
	return nil
}

// clearLocalTierCache removes current-process entries after a local or shared revision change.
func (s *serviceImpl) clearLocalTierCache(revision int64, capabilityType string, capabilityMethod string, tierCode string) {
	if s == nil {
		return
	}
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	if capabilityType == "" && capabilityMethod == "" && tierCode == "" {
		s.tierCache = make(map[string]tierCacheEntry)
		s.revision = revision
		return
	}
	delete(s.tierCache, tierCacheKey(
		normalizeCapabilityType(capabilityType),
		normalizeCapabilityMethod(capabilityMethod),
		normalizeTierCode(tierCode),
	))
	s.revision = revision
}

// cacheRevisionValue extracts the integer revision published through host cache.
func cacheRevisionValue(item *cachecap.CacheItem) int64 {
	if item == nil {
		return 0
	}
	return item.IntValue
}

// resolveTierBinding returns one cached or database-loaded binding.
func (s *serviceImpl) resolveTierBinding(ctx context.Context, capabilityType string, capabilityMethod string, tierCode string) (*resolvedTierBinding, error) {
	capabilityType = normalizeCapabilityType(capabilityType)
	capabilityMethod = normalizeCapabilityMethod(capabilityMethod)
	tierCode = normalizeTierCode(tierCode)
	if !s.tierCacheEnabled(ctx) {
		return s.loadTierBinding(ctx, capabilityType, capabilityMethod, tierCode)
	}
	if err := s.observeTierCacheRevision(ctx); err != nil {
		return nil, err
	}
	key := tierCacheKey(capabilityType, capabilityMethod, tierCode)
	now := time.Now()
	s.cacheMu.RLock()
	if cached, ok := s.tierCache[key]; ok && now.Before(cached.expiresAt) && cached.value != nil {
		value := *cached.value
		s.cacheMu.RUnlock()
		return &value, nil
	}
	s.cacheMu.RUnlock()
	value, err := s.loadTierBinding(ctx, capabilityType, capabilityMethod, tierCode)
	if err != nil {
		return nil, err
	}
	s.cacheMu.Lock()
	s.tierCache[key] = tierCacheEntry{
		value:     value,
		expiresAt: now.Add(tierCacheTTL),
	}
	s.cacheMu.Unlock()
	return value, nil
}

// tierCacheEnabled reports whether this request can safely use the platform-level tier cache.
func (s *serviceImpl) tierCacheEnabled(ctx context.Context) bool {
	if s == nil {
		return false
	}
	current := bizctxcap.CurrentFromContext(ctx)
	if s.bizCtxSvc != nil {
		current = s.bizCtxSvc.Current(ctx)
	}
	return current.TenantID == 0
}

// loadTierBinding loads one tier binding from database.
func (s *serviceImpl) loadTierBinding(ctx context.Context, capabilityType string, capabilityMethod string, tierCode string) (*resolvedTierBinding, error) {
	tier, err := s.getTier(ctx, capabilityType, capabilityMethod, tierCode)
	if err != nil {
		return nil, err
	}
	if tier.Enabled != enabledYes {
		return nil, bizerr.NewCode(CodeTierBindingUnavailable)
	}
	bindings, err := s.primaryBindingsByTier(ctx, []int64{tier.Id})
	if err != nil {
		return nil, err
	}
	binding := bindings[tier.Id]
	if binding == nil {
		return nil, bizerr.NewCode(CodeTierBindingUnavailable)
	}
	return binding, nil
}

// getTier returns one fixed tier by capability method and code.
func (s *serviceImpl) getTier(ctx context.Context, capabilityType string, capabilityMethod string, tierCode string) (*entity.Tier, error) {
	var row *entity.Tier
	if err := dao.Tier.Ctx(ctx).
		Where(do.Tier{
			CapabilityType:   normalizeCapabilityType(capabilityType),
			CapabilityMethod: normalizeCapabilityMethod(capabilityMethod),
			Code:             normalizeTierCode(tierCode),
		}).
		Scan(&row); err != nil {
		return nil, err
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeTierNotFound)
	}
	return row, nil
}

// validateModelBinding verifies provider, model, and endpoint existence.
func (s *serviceImpl) validateModelBinding(
	ctx context.Context,
	providerID int64,
	modelID int64,
) (*entity.Model, *entity.ProviderEndpoint, error) {
	provider, err := s.getProvider(ctx, providerID)
	if err != nil {
		return nil, nil, err
	}
	if provider.Enabled != enabledYes {
		return nil, nil, bizerr.NewCode(CodeProviderNotFound)
	}
	model, err := s.getModel(ctx, modelID)
	if err != nil {
		return nil, nil, err
	}
	if model.ProviderId != providerID || model.Enabled != enabledYes {
		return nil, nil, bizerr.NewCode(CodeModelNotFound)
	}
	endpoint, err := s.requireProviderEndpoint(ctx, providerID, model.EndpointId, model.Protocol)
	if err != nil {
		return nil, nil, err
	}
	return model, endpoint, nil
}

// validateModelCapabilityBinding requires explicit model capability declarations
// for multimodal methods while keeping text.generate declaration-free for the
// framework default path.
func (s *serviceImpl) validateModelCapabilityBinding(
	ctx context.Context,
	modelID int64,
	capabilityType string,
	capabilityMethod string,
) error {
	capabilityType = normalizeCapabilityType(capabilityType)
	capabilityMethod = normalizeCapabilityMethod(capabilityMethod)
	if capabilityType == CapabilityTypeText && capabilityMethod == CapabilityMethodGenerate {
		return nil
	}
	count, err := dao.ModelCapability.Ctx(ctx).Where(do.ModelCapability{
		ModelId:          modelID,
		CapabilityType:   capabilityType,
		CapabilityMethod: capabilityMethod,
		Enabled:          enabledYes,
	}).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return bizerr.NewCode(CodeModelNotFound)
	}
	return nil
}

// upsertPrimaryBinding inserts or updates the active primary binding row.
func (s *serviceImpl) upsertPrimaryBinding(ctx context.Context, tierID int64, providerID int64, modelID int64) error {
	var existing *entity.TierBinding
	if err := dao.TierBinding.Ctx(ctx).
		Where(do.TierBinding{TierId: tierID, Priority: primaryBindingPriority}).
		Scan(&existing); err != nil {
		return err
	}
	data := do.TierBinding{
		TierId:     tierID,
		ProviderId: providerID,
		ModelId:    modelID,
		Priority:   primaryBindingPriority,
		Enabled:    enabledYes,
	}
	if existing == nil {
		_, err := dao.TierBinding.Ctx(ctx).Data(data).Insert()
		return err
	}
	_, err := dao.TierBinding.Ctx(ctx).Where(do.TierBinding{Id: existing.Id}).Data(data).Update()
	return err
}

// primaryBindingsByTier resolves primary bindings and associated providers/models in batches.
func (s *serviceImpl) primaryBindingsByTier(ctx context.Context, tierIDs []int64) (map[int64]*resolvedTierBinding, error) {
	result := make(map[int64]*resolvedTierBinding, len(tierIDs))
	if len(tierIDs) == 0 {
		return result, nil
	}
	bindingRows := make([]*entity.TierBinding, 0)
	if err := dao.TierBinding.Ctx(ctx).
		WhereIn(dao.TierBinding.Columns().TierId, tierIDs).
		Where(do.TierBinding{Priority: primaryBindingPriority, Enabled: enabledYes}).
		Scan(&bindingRows); err != nil {
		return nil, err
	}
	providerIDs := make([]int64, 0, len(bindingRows))
	modelIDs := make([]int64, 0, len(bindingRows))
	for _, row := range bindingRows {
		providerIDs = append(providerIDs, row.ProviderId)
		modelIDs = append(modelIDs, row.ModelId)
	}
	providers, err := s.providersByID(ctx, providerIDs)
	if err != nil {
		return nil, err
	}
	models, err := s.modelsByID(ctx, modelIDs)
	if err != nil {
		return nil, err
	}
	tiers, err := s.tiersByID(ctx, tierIDs)
	if err != nil {
		return nil, err
	}
	endpointIDs := make([]int64, 0, len(models))
	for _, row := range bindingRows {
		tier := tiers[row.TierId]
		model := models[row.ModelId]
		if tier == nil || model == nil {
			continue
		}
		endpointID := model.EndpointId
		if endpointID > 0 {
			endpointIDs = append(endpointIDs, endpointID)
		}
	}
	endpoints, err := s.endpointsByID(ctx, endpointIDs)
	if err != nil {
		return nil, err
	}
	for _, row := range bindingRows {
		tier := tiers[row.TierId]
		provider := providers[row.ProviderId]
		model := models[row.ModelId]
		if tier == nil || provider == nil || model == nil || provider.Enabled != enabledYes || model.Enabled != enabledYes {
			continue
		}
		endpointID := model.EndpointId
		endpoint := endpoints[endpointID]
		if endpoint == nil || endpoint.ProviderId != provider.Id || endpoint.Enabled != enabledYes || endpoint.Protocol != model.Protocol {
			continue
		}
		result[row.TierId] = resolvedBindingFromRows(tier, provider, model, endpoint)
	}
	return result, nil
}

// resolvedBindingFromRows builds a resolved binding from loaded rows.
func resolvedBindingFromRows(
	tier *entity.Tier,
	provider *entity.Provider,
	model *entity.Model,
	endpoint *entity.ProviderEndpoint,
) *resolvedTierBinding {
	return &resolvedTierBinding{
		TierId:            tier.Id,
		TierCode:          tier.Code,
		CapabilityType:    tier.CapabilityType,
		CapabilityMethod:  tier.CapabilityMethod,
		DefaultEffort:     tier.DefaultEffort,
		ProviderId:        provider.Id,
		ProviderName:      provider.Name,
		ModelId:           model.Id,
		ModelName:         model.ModelName,
		Protocol:          model.Protocol,
		EndpointId:        endpoint.Id,
		EndpointBaseUrl:   endpoint.BaseUrl,
		EndpointSecretRef: endpoint.SecretRef,
	}
}

// providersByID loads providers by ID in one query.
func (s *serviceImpl) providersByID(ctx context.Context, ids []int64) (map[int64]*entity.Provider, error) {
	out := make(map[int64]*entity.Provider, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	rows := make([]*entity.Provider, 0)
	if err := dao.Provider.Ctx(ctx).WhereIn(dao.Provider.Columns().Id, ids).Scan(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.Id] = row
	}
	return out, nil
}

// modelsByID loads models by ID in one query.
func (s *serviceImpl) modelsByID(ctx context.Context, ids []int64) (map[int64]*entity.Model, error) {
	out := make(map[int64]*entity.Model, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	rows := make([]*entity.Model, 0)
	if err := dao.Model.Ctx(ctx).WhereIn(dao.Model.Columns().Id, ids).Scan(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.Id] = row
	}
	return out, nil
}

// endpointsByID loads provider endpoints by ID in one query.
func (s *serviceImpl) endpointsByID(ctx context.Context, ids []int64) (map[int64]*entity.ProviderEndpoint, error) {
	out := make(map[int64]*entity.ProviderEndpoint, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	rows := make([]*entity.ProviderEndpoint, 0)
	if err := dao.ProviderEndpoint.Ctx(ctx).WhereIn(dao.ProviderEndpoint.Columns().Id, ids).Scan(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.Id] = row
	}
	return out, nil
}

// tiersByID loads tiers by ID in one query.
func (s *serviceImpl) tiersByID(ctx context.Context, ids []int64) (map[int64]*entity.Tier, error) {
	out := make(map[int64]*entity.Tier, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	rows := make([]*entity.Tier, 0)
	if err := dao.Tier.Ctx(ctx).WhereIn(dao.Tier.Columns().Id, ids).Scan(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.Id] = row
	}
	return out, nil
}

// collectTierIDs extracts tier IDs.
func collectTierIDs(rows []*entity.Tier) []int64 {
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row != nil && row.Id > 0 {
			ids = append(ids, row.Id)
		}
	}
	return ids
}

// tierCacheKey builds the cache key for one capability method and tier pair.
func tierCacheKey(capabilityType string, capabilityMethod string, tierCode string) string {
	return normalizeCapabilityType(capabilityType) + ":" +
		normalizeCapabilityMethod(capabilityMethod) + ":" +
		normalizeTierCode(tierCode)
}

// effortSupportedByBinding only validates platform effort enum syntax. Model
// support is intentionally not inferred from model management metadata.
func effortSupportedByBinding(binding *resolvedTierBinding, effort string) bool {
	normalized, err := normalizeEffort(effort)
	return binding != nil && err == nil && (normalized != "" || effort == "")
}

// defaultTierTestMessages returns a minimal connectivity test prompt.
func defaultTierTestMessages() []aitext.Message {
	return []aitext.Message{{Role: aitext.MessageRoleUser, Content: "Reply with OK."}}
}
