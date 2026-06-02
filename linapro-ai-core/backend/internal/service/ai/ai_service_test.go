// This file verifies Smart Center service behavior against plugin-owned tables.

package ai

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"

	"lina-core/pkg/bizerr"
	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/plugin/capability/ai/aitext"
	plugincontract "lina-core/pkg/plugin/capability/contract"
	"lina-plugin-linapro-ai-core/backend/internal/dao"
	"lina-plugin-linapro-ai-core/backend/internal/model/do"
	"lina-plugin-linapro-ai-core/backend/internal/model/entity"
)

var (
	installSQLOnce  sync.Once
	installSQLError error
)

// TestProviderModelTierAndInvocationFlow verifies provider/model CRUD,
// reference protection, tier save, cache invalidation, text generation, and
// masked invocation query behavior in one isolated provider fixture.
func TestProviderModelTierAndInvocationFlow(t *testing.T) {
	ctx := context.Background()
	prepareDatabase(t, ctx)
	server := testOpenAIServer(t)
	ctx = context.WithValue(ctx, providerBaseURLKey{}, server.URL+"/v1")
	svc := New(testBizCtx{}, nil, server.Client()).(*serviceImpl)
	snapshot := snapshotTier(t, ctx, svc, TierCodeBasic)
	t.Cleanup(func() { restoreTier(t, ctx, snapshot) })

	providerID := createTestProvider(t, ctx, svc, "flow")
	t.Cleanup(func() { deleteProviderFixture(t, ctx, providerID) })
	modelID := createTestModel(t, ctx, svc, providerID, "flow-model")

	providers, err := svc.ListProviders(ctx, ProviderListInput{PageNum: 1, PageSize: 10, Keyword: "flow"})
	if err != nil {
		t.Fatalf("list providers: %v", err)
	}
	if providers.Total < 1 || providers.List[0].ModelCount != 1 || providers.List[0].EnabledModelCount != 1 {
		t.Fatalf("expected aggregated model counts, got %#v", providers)
	}

	if err = svc.UpdateTier(ctx, TierUpdateInput{
		Code:          TierCodeBasic,
		ProviderId:    providerID,
		ModelId:       modelID,
		DefaultEffort: string(aitext.ThinkingEffortLow),
		Enabled:       enabledYes,
	}); err != nil {
		t.Fatalf("update tier: %v", err)
	}
	if err = svc.UpdateTier(ctx, TierUpdateInput{
		Code:          TierCodeBasic,
		DefaultEffort: string(aitext.ThinkingEffortLow),
		Enabled:       enabledNo,
	}); err != nil {
		t.Fatalf("disable tier without rebinding: %v", err)
	}
	preserved := snapshotTier(t, ctx, svc, TierCodeBasic)
	if preserved.binding == nil || preserved.binding.ProviderId != providerID || preserved.binding.ModelId != modelID {
		t.Fatalf("expected disabling tier to preserve binding, got %#v", preserved.binding)
	}
	if err = svc.UpdateTier(ctx, TierUpdateInput{
		Code:          TierCodeBasic,
		ProviderId:    providerID,
		ModelId:       modelID,
		DefaultEffort: string(aitext.ThinkingEffortLow),
		Enabled:       enabledYes,
	}); err != nil {
		t.Fatalf("re-enable tier: %v", err)
	}
	syncOut, err := svc.SyncModels(ctx, ModelSyncInput{ProviderId: providerID, Protocol: ProtocolOpenAI})
	if err != nil {
		t.Fatalf("sync models: %v", err)
	}
	if syncOut.Created != 1 || syncOut.Kept != 1 {
		t.Fatalf("expected sync to batch-detect existing models, got %#v", syncOut)
	}
	if err = svc.DeleteModel(ctx, modelID); !bizerr.Is(err, CodeModelInUse) {
		t.Fatalf("expected model in-use protection, got %v", err)
	}
	if err = svc.DeleteProvider(ctx, providerID); !bizerr.Is(err, CodeProviderInUse) {
		t.Fatalf("expected provider in-use protection, got %v", err)
	}

	response, err := svc.GenerateText(ctx, aitext.ProviderRequest{
		GenerateRequest: aitext.GenerateRequest{
			Purpose: "unit.flow",
			Tier:    aitext.TierBasic,
			Messages: []aitext.Message{
				{Role: aitext.MessageRoleUser, Content: "full prompt must not be logged"},
			},
			MaxOutputTokens: 64,
		},
	})
	if err != nil {
		t.Fatalf("generate text: %v", err)
	}
	if response.Text != "provider ok" || response.ModelName != "flow-model" {
		t.Fatalf("unexpected response: %#v", response)
	}

	logs, err := svc.ListInvocations(ctx, InvocationListInput{
		PageNum:  1,
		PageSize: 10,
		Purpose:  "unit.flow",
		Status:   InvocationStatusSuccess,
	})
	if err != nil {
		t.Fatalf("list invocations: %v", err)
	}
	if logs.Total < 1 || strings.Contains(logs.List[0].ErrorSummary, "full prompt") {
		t.Fatalf("expected masked invocation log, got %#v", logs)
	}
}

// TestTierDraftTestDoesNotPersistBinding verifies draft tier testing calls the
// provider but does not persist the draft binding as the primary tier binding.
func TestTierDraftTestDoesNotPersistBinding(t *testing.T) {
	ctx := context.Background()
	prepareDatabase(t, ctx)
	server := testOpenAIServer(t)
	ctx = context.WithValue(ctx, providerBaseURLKey{}, server.URL+"/v1")
	svc := New(testBizCtx{}, nil, server.Client()).(*serviceImpl)
	snapshot := snapshotTier(t, ctx, svc, TierCodeStandard)
	t.Cleanup(func() { restoreTier(t, ctx, snapshot) })

	providerID := createTestProvider(t, ctx, svc, "draft")
	t.Cleanup(func() { deleteProviderFixture(t, ctx, providerID) })
	modelID := createTestModel(t, ctx, svc, providerID, "draft-model")

	out, err := svc.TestTier(ctx, TierTestInput{
		Code:           TierCodeStandard,
		ProviderId:     providerID,
		ModelId:        modelID,
		ThinkingEffort: string(aitext.ThinkingEffortLow),
		Messages:       []aitext.Message{{Role: aitext.MessageRoleUser, Content: "ping"}},
	})
	if err != nil {
		t.Fatalf("test draft tier: %v", err)
	}
	if out.Status != InvocationStatusSuccess || out.ModelName != "draft-model" {
		t.Fatalf("unexpected draft test output: %#v", out)
	}
	count, err := dao.TierBinding.Ctx(ctx).Where(do.TierBinding{
		ProviderId: providerID,
		ModelId:    modelID,
		Priority:   primaryBindingPriority,
	}).Count()
	if err != nil {
		t.Fatalf("count draft binding: %v", err)
	}
	if count != 0 {
		t.Fatalf("draft tier test persisted binding count=%d", count)
	}
}

// TestTierCacheSharedRevisionInvalidatesPeerService verifies that one service
// instance observes another instance's tier-binding mutation through the
// plugin-scoped shared cache revision before using its local tier cache.
func TestTierCacheSharedRevisionInvalidatesPeerService(t *testing.T) {
	ctx := context.Background()
	prepareDatabase(t, ctx)
	server := testOpenAIServer(t)
	ctx = context.WithValue(ctx, providerBaseURLKey{}, server.URL+"/v1")
	cacheSvc := newMemoryCacheService()
	writer := New(testBizCtx{}, cacheSvc, server.Client()).(*serviceImpl)
	reader := New(testBizCtx{}, cacheSvc, server.Client()).(*serviceImpl)
	snapshot := snapshotTier(t, ctx, writer, TierCodeAdvanced)
	t.Cleanup(func() { restoreTier(t, ctx, snapshot) })

	firstProviderID := createTestProvider(t, ctx, writer, "revision-first")
	t.Cleanup(func() { deleteProviderFixture(t, ctx, firstProviderID) })
	firstModelID := createTestModel(t, ctx, writer, firstProviderID, "revision-first-model")
	if err := writer.UpdateTier(ctx, TierUpdateInput{
		Code:          TierCodeAdvanced,
		ProviderId:    firstProviderID,
		ModelId:       firstModelID,
		DefaultEffort: string(aitext.ThinkingEffortLow),
		Enabled:       enabledYes,
	}); err != nil {
		t.Fatalf("bind first tier: %v", err)
	}
	firstBinding, err := reader.resolveTierBinding(ctx, CapabilityTypeText, TierCodeAdvanced)
	if err != nil {
		t.Fatalf("resolve first binding: %v", err)
	}
	if firstBinding.ModelName != "revision-first-model" {
		t.Fatalf("expected first model, got %#v", firstBinding)
	}

	secondProviderID := createTestProvider(t, ctx, writer, "revision-second")
	t.Cleanup(func() { deleteProviderFixture(t, ctx, secondProviderID) })
	secondModelID := createTestModel(t, ctx, writer, secondProviderID, "revision-second-model")
	if err = writer.UpdateTier(ctx, TierUpdateInput{
		Code:          TierCodeAdvanced,
		ProviderId:    secondProviderID,
		ModelId:       secondModelID,
		DefaultEffort: string(aitext.ThinkingEffortLow),
		Enabled:       enabledYes,
	}); err != nil {
		t.Fatalf("bind second tier: %v", err)
	}
	secondBinding, err := reader.resolveTierBinding(ctx, CapabilityTypeText, TierCodeAdvanced)
	if err != nil {
		t.Fatalf("resolve second binding: %v", err)
	}
	if secondBinding.ModelName != "revision-second-model" {
		t.Fatalf("expected peer cache invalidation to load second model, got %#v", secondBinding)
	}
}

type testBizCtx struct{}

func (testBizCtx) Current(context.Context) plugincontract.CurrentContext {
	return plugincontract.CurrentContext{UserID: 1, TenantID: 0, PlatformBypass: true}
}

type memoryCacheService struct {
	mu    sync.Mutex
	items map[string]*plugincontract.CacheItem
}

func newMemoryCacheService() *memoryCacheService {
	return &memoryCacheService{items: make(map[string]*plugincontract.CacheItem)}
}

func (s *memoryCacheService) Get(
	_ context.Context,
	namespace string,
	key string,
) (*plugincontract.CacheItem, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[memoryCacheKey(namespace, key)]
	if !ok {
		return nil, false, nil
	}
	copied := *item
	return &copied, true, nil
}

func (s *memoryCacheService) Set(
	_ context.Context,
	namespace string,
	key string,
	value string,
	ttl time.Duration,
) (*plugincontract.CacheItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item := &plugincontract.CacheItem{
		Key:       key,
		ValueKind: plugincontract.CacheValueKindString,
		Value:     value,
		ExpireAt:  memoryCacheExpireAt(ttl),
	}
	s.items[memoryCacheKey(namespace, key)] = item
	copied := *item
	return &copied, nil
}

func (s *memoryCacheService) Delete(_ context.Context, namespace string, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, memoryCacheKey(namespace, key))
	return nil
}

func (s *memoryCacheService) Incr(
	_ context.Context,
	namespace string,
	key string,
	delta int64,
	ttl time.Duration,
) (*plugincontract.CacheItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cacheKey := memoryCacheKey(namespace, key)
	item := s.items[cacheKey]
	if item == nil {
		item = &plugincontract.CacheItem{Key: key, ValueKind: plugincontract.CacheValueKindInt}
		s.items[cacheKey] = item
	}
	item.IntValue += delta
	if ttl > 0 {
		item.ExpireAt = memoryCacheExpireAt(ttl)
	}
	copied := *item
	return &copied, nil
}

func (s *memoryCacheService) Expire(
	_ context.Context,
	namespace string,
	key string,
	ttl time.Duration,
) (bool, *time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item := s.items[memoryCacheKey(namespace, key)]
	if item == nil {
		return false, nil, nil
	}
	item.ExpireAt = memoryCacheExpireAt(ttl)
	return true, item.ExpireAt, nil
}

func memoryCacheKey(namespace string, key string) string {
	return namespace + "\x00" + key
}

func memoryCacheExpireAt(ttl time.Duration) *time.Time {
	if ttl <= 0 {
		return nil
	}
	expireAt := time.Now().Add(ttl)
	return &expireAt
}

type tierSnapshot struct {
	tier    *entity.Tier
	binding *entity.TierBinding
}

func prepareDatabase(t *testing.T, ctx context.Context) {
	t.Helper()
	adapter, err := gcfg.NewAdapterContent(`
database:
  default:
    link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
    debug: false
`)
	if err != nil {
		t.Fatalf("create config adapter: %v", err)
	}
	g.Cfg().SetAdapter(adapter)

	installSQLOnce.Do(func() {
		sqlPath := filepath.Clean("../../../../manifest/sql/001-linapro-ai-core-schema.sql")
		content, readErr := os.ReadFile(sqlPath)
		if readErr != nil {
			installSQLError = readErr
			return
		}
		for _, statement := range strings.Split(string(content), ";") {
			if strings.TrimSpace(statement) == "" {
				continue
			}
			if _, execErr := g.DB().Exec(ctx, statement); execErr != nil {
				installSQLError = execErr
				return
			}
		}
	})
	if installSQLError != nil {
		t.Skipf("database unavailable for linapro-ai-core service test: %v", installSQLError)
	}
	if _, err = g.DB().Exec(ctx, "SELECT 1"); err != nil {
		t.Skipf("database unavailable for linapro-ai-core service test: %v", err)
	}
}

func createTestProvider(t *testing.T, ctx context.Context, svc *serviceImpl, suffix string) int64 {
	t.Helper()
	id, err := svc.CreateProvider(ctx, ProviderSaveInput{
		Name:            "unit-provider-" + suffix + "-" + time.Now().Format("150405.000000000"),
		OpenaiBaseUrl:   testProviderBaseURL(ctx),
		ApiKeySecretRef: "unit-secret",
		Enabled:         enabledYes,
	})
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}
	return id
}

func createTestModel(t *testing.T, ctx context.Context, svc *serviceImpl, providerID int64, modelName string) int64 {
	t.Helper()
	id, err := svc.CreateModel(ctx, ModelSaveInput{
		ProviderId:       providerID,
		CapabilityType:   CapabilityTypeText,
		ModelName:        modelName,
		Protocol:         ProtocolOpenAI,
		SupportsThinking: enabledYes,
		SupportedEfforts: []string{string(aitext.ThinkingEffortLow), string(aitext.ThinkingEffortMedium)},
		MaxInputTokens:   1024,
		MaxOutputTokens:  256,
		Enabled:          enabledYes,
	})
	if err != nil {
		t.Fatalf("create model: %v", err)
	}
	return id
}

func deleteProviderFixture(t *testing.T, ctx context.Context, providerID int64) {
	t.Helper()
	statements := []string{
		`DELETE FROM plugin_linapro_ai_invocation WHERE provider_id = $1`,
		`DELETE FROM plugin_linapro_ai_tier_binding WHERE provider_id = $1`,
		`DELETE FROM plugin_linapro_ai_model WHERE provider_id = $1`,
		`DELETE FROM plugin_linapro_ai_provider WHERE id = $1`,
	}
	for _, statement := range statements {
		if _, err := g.DB().Exec(ctx, statement, providerID); err != nil {
			t.Fatalf("cleanup provider fixture: %v", err)
		}
	}
}

func snapshotTier(t *testing.T, ctx context.Context, svc *serviceImpl, code string) tierSnapshot {
	t.Helper()
	tier, err := svc.getTier(ctx, CapabilityTypeText, code)
	if err != nil {
		t.Fatalf("snapshot tier: %v", err)
	}
	var binding *entity.TierBinding
	if err = dao.TierBinding.Ctx(ctx).Where(do.TierBinding{
		TierId:   tier.Id,
		Priority: primaryBindingPriority,
	}).Scan(&binding); err != nil {
		t.Fatalf("snapshot tier binding: %v", err)
	}
	return tierSnapshot{tier: tier, binding: binding}
}

func restoreTier(t *testing.T, ctx context.Context, snapshot tierSnapshot) {
	t.Helper()
	if snapshot.tier == nil {
		return
	}
	if _, err := dao.Tier.Ctx(ctx).Where(do.Tier{Id: snapshot.tier.Id}).Data(do.Tier{
		DefaultEffort:        snapshot.tier.DefaultEffort,
		Enabled:              snapshot.tier.Enabled,
		LastTestStatus:       snapshot.tier.LastTestStatus,
		LastTestLatencyMs:    snapshot.tier.LastTestLatencyMs,
		LastTestErrorSummary: snapshot.tier.LastTestErrorSummary,
		LastTestAt:           snapshot.tier.LastTestAt,
	}).Update(); err != nil {
		t.Fatalf("restore tier: %v", err)
	}
	if _, err := dao.TierBinding.Ctx(ctx).Where(do.TierBinding{
		TierId:   snapshot.tier.Id,
		Priority: primaryBindingPriority,
	}).Delete(); err != nil {
		t.Fatalf("clear tier binding: %v", err)
	}
	if snapshot.binding == nil {
		return
	}
	if _, err := dao.TierBinding.Ctx(ctx).Data(do.TierBinding{
		TierId:     snapshot.binding.TierId,
		ProviderId: snapshot.binding.ProviderId,
		ModelId:    snapshot.binding.ModelId,
		Priority:   snapshot.binding.Priority,
		Enabled:    snapshot.binding.Enabled,
	}).Insert(); err != nil {
		t.Fatalf("restore tier binding: %v", err)
	}
}
