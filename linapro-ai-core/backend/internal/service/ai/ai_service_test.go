// This file verifies Smart Center service behavior against plugin-owned tables.

package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
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
	if len(providers.List[0].Models) != 1 || providers.List[0].Models[0].ModelName != "flow-model" {
		t.Fatalf("expected batched model summaries, got %#v", providers.List[0].Models)
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
	if err = svc.UpdateModel(ctx, ModelSaveInput{
		Id:               modelID,
		EndpointId:       defaultTestEndpointID(t, ctx, svc, providerID, ProtocolOpenAI),
		CapabilityType:   "image",
		CapabilityMethod: "generate",
		ModelName:        "flow-model-renamed",
		Protocol:         ProtocolOpenAI,
		SupportsThinking: enabledNo,
		SupportedEfforts: []string{string(aitext.ThinkingEffortHigh)},
		MaxInputTokens:   999,
		MaxOutputTokens:  999,
		Enabled:          enabledYes,
	}); err != nil {
		t.Fatalf("update model identity: %v", err)
	}
	capabilities, err := svc.ListModelCapabilities(ctx, modelID)
	if err != nil {
		t.Fatalf("list capabilities after model update: %v", err)
	}
	if len(capabilities) != 1 ||
		capabilities[0].CapabilityType != CapabilityTypeText ||
		capabilities[0].CapabilityMethod != CapabilityMethodGenerate ||
		capabilities[0].SupportsThinking != enabledYes ||
		capabilities[0].MaxOutputTokens != 256 {
		t.Fatalf("model update must not mutate capability declarations, got %#v", capabilities)
	}
	models, err := svc.ListModels(ctx, ModelListInput{
		ProviderId:       providerID,
		PageNum:          1,
		PageSize:         10,
		CapabilityType:   CapabilityTypeText,
		CapabilityMethod: CapabilityMethodGenerate,
		Enabled:          intPtr(enabledYes),
	})
	if err != nil {
		t.Fatalf("list models after model update: %v", err)
	}
	if models.Total != 2 ||
		models.List[0].ModelName != "flow-model-renamed" ||
		models.List[0].CapabilityType != CapabilityTypeText ||
		models.List[0].SupportsThinking != enabledYes {
		t.Fatalf("expected list model projection from capability table, got %#v", models)
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
	if response.Text != "provider ok" || response.ModelName != "flow-model-renamed" {
		t.Fatalf("unexpected response: %#v", response)
	}

	logs, err := svc.ListInvocations(ctx, InvocationListInput{
		PageNum:          1,
		PageSize:         10,
		CapabilityMethod: CapabilityMethodGenerate,
		Purpose:          "unit.flow",
		Status:           InvocationStatusSuccess,
	})
	if err != nil {
		t.Fatalf("list invocations: %v", err)
	}
	if logs.Total < 1 || logs.List[0].CapabilityMethod != CapabilityMethodGenerate || strings.Contains(logs.List[0].ErrorSummary, "full prompt") {
		t.Fatalf("expected masked invocation log, got %#v", logs)
	}
}

// TestSyncModelsAggregatesEnabledEndpoints verifies provider sync can import
// from successful endpoints while tolerating unsupported or failing endpoints,
// but still returns an error when every endpoint fails.
func TestSyncModelsAggregatesEnabledEndpoints(t *testing.T) {
	ctx := context.Background()
	prepareDatabase(t, ctx)
	openAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" && r.URL.Path != "/models" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"data":[{"id":"sync-existing"},{"id":"sync-created"}]}`)); err != nil {
			t.Fatalf("write OpenAI sync response: %v", err)
		}
	}))
	t.Cleanup(openAIServer.Close)
	failingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	t.Cleanup(failingServer.Close)

	ctx = context.WithValue(ctx, providerBaseURLKey{}, openAIServer.URL+"/v1")
	svc := New(testBizCtx{}, nil, openAIServer.Client()).(*serviceImpl)
	providerID := createTestProvider(t, ctx, svc, "aggregate-sync")
	t.Cleanup(func() { deleteProviderFixture(t, ctx, providerID) })
	createTestModel(t, ctx, svc, providerID, "sync-existing")
	if _, err := svc.CreateProviderEndpoint(ctx, ProviderEndpointSaveInput{
		ProviderId: providerID,
		Protocol:   ProtocolAnthropic,
		BaseUrl:    failingServer.URL + "/v1",
		SecretRef:  "unit-secret",
		Enabled:    enabledYes,
	}); err != nil {
		t.Fatalf("create failing endpoint: %v", err)
	}

	out, err := svc.SyncModels(ctx, ModelSyncInput{ProviderId: providerID})
	if err != nil {
		t.Fatalf("sync aggregate endpoints: %v", err)
	}
	if out.Created != 1 || out.Kept != 1 {
		t.Fatalf("expected one created and one kept model, got %#v", out)
	}
	models, err := svc.ListAllModels(ctx, ModelGlobalListInput{
		Keyword:  "sync-created",
		PageNum:  1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("list synced model: %v", err)
	}
	if models.Total != 1 || len(models.List) != 1 {
		t.Fatalf("expected one synced model projection, got %#v", models)
	}
	if models.List[0].ProviderName == "" ||
		models.List[0].EndpointBaseUrl != openAIServer.URL+"/v1" ||
		models.List[0].Protocol != ProtocolOpenAI {
		t.Fatalf("expected synced model provider and endpoint projection, got %#v", models.List[0])
	}

	failingProviderID, err := svc.CreateProvider(ctx, ProviderSaveInput{
		Name:    "unit-provider-all-failing-sync-" + time.Now().Format("150405.000000000"),
		Enabled: enabledYes,
	})
	if err != nil {
		t.Fatalf("create all-failing provider: %v", err)
	}
	t.Cleanup(func() { deleteProviderFixture(t, ctx, failingProviderID) })
	for _, protocol := range []string{ProtocolOpenAI, ProtocolAnthropic} {
		if _, err = svc.CreateProviderEndpoint(ctx, ProviderEndpointSaveInput{
			ProviderId: failingProviderID,
			Protocol:   protocol,
			BaseUrl:    failingServer.URL + "/v1",
			SecretRef:  "unit-secret",
			Enabled:    enabledYes,
		}); err != nil {
			t.Fatalf("create all-failing endpoint %s: %v", protocol, err)
		}
	}
	if _, err = svc.SyncModels(ctx, ModelSyncInput{ProviderId: failingProviderID}); err == nil {
		t.Fatalf("expected all-failing sync to return an error")
	}
}

// TestTierBindingRequiresMatchingCapabilityMethod verifies text.generate tiers
// cannot bind provider models declared for another method in the same type.
func TestTierBindingRequiresMatchingCapabilityMethod(t *testing.T) {
	ctx := context.Background()
	prepareDatabase(t, ctx)
	server := testOpenAIServer(t)
	ctx = context.WithValue(ctx, providerBaseURLKey{}, server.URL+"/v1")
	svc := New(testBizCtx{}, nil, server.Client()).(*serviceImpl)
	snapshot := snapshotTier(t, ctx, svc, TierCodeBasic)
	t.Cleanup(func() { restoreTier(t, ctx, snapshot) })

	providerID := createTestProvider(t, ctx, svc, "wrong-method")
	t.Cleanup(func() { deleteProviderFixture(t, ctx, providerID) })
	modelID, err := svc.CreateModel(ctx, ModelSaveInput{
		ProviderId:       providerID,
		EndpointId:       defaultTestEndpointID(t, ctx, svc, providerID, ProtocolOpenAI),
		CapabilityType:   CapabilityTypeText,
		CapabilityMethod: "summarize",
		ModelName:        "wrong-method-model",
		Protocol:         ProtocolOpenAI,
		SupportsThinking: enabledYes,
		SupportedEfforts: []string{string(aitext.ThinkingEffortLow)},
		MaxInputTokens:   1024,
		MaxOutputTokens:  256,
		Enabled:          enabledYes,
	})
	if err != nil {
		t.Fatalf("create wrong-method model: %v", err)
	}

	err = svc.UpdateTier(ctx, TierUpdateInput{
		CapabilityType:   CapabilityTypeText,
		CapabilityMethod: CapabilityMethodGenerate,
		Code:             TierCodeBasic,
		ProviderId:       providerID,
		ModelId:          modelID,
		DefaultEffort:    string(aitext.ThinkingEffortLow),
		Enabled:          enabledYes,
	})
	if !bizerr.Is(err, CodeModelNotFound) {
		t.Fatalf("expected method mismatch model to be rejected, got %v", err)
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
	afterDraft := snapshotTier(t, ctx, svc, TierCodeStandard)
	if afterDraft.tier.LastTestStatus != snapshot.tier.LastTestStatus ||
		afterDraft.tier.LastTestLatencyMs != snapshot.tier.LastTestLatencyMs ||
		afterDraft.tier.LastTestErrorSummary != snapshot.tier.LastTestErrorSummary ||
		!sameOptionalTime(afterDraft.tier.LastTestAt, snapshot.tier.LastTestAt) {
		t.Fatalf("draft tier test changed last test summary, before=%#v after=%#v", snapshot.tier, afterDraft.tier)
	}

	if err = svc.UpdateTier(ctx, TierUpdateInput{
		Code:          TierCodeStandard,
		ProviderId:    providerID,
		ModelId:       modelID,
		DefaultEffort: string(aitext.ThinkingEffortLow),
		Enabled:       enabledYes,
	}); err != nil {
		t.Fatalf("save tier binding: %v", err)
	}
	beforeSavedTest := snapshotTier(t, ctx, svc, TierCodeStandard)
	savedOut, err := svc.TestTier(ctx, TierTestInput{
		Code:     TierCodeStandard,
		Messages: []aitext.Message{{Role: aitext.MessageRoleUser, Content: "ping"}},
	})
	if err != nil {
		t.Fatalf("test saved tier: %v", err)
	}
	if savedOut.Status != InvocationStatusSuccess || savedOut.ModelName != "draft-model" {
		t.Fatalf("unexpected saved test output: %#v", savedOut)
	}
	afterSavedTest := snapshotTier(t, ctx, svc, TierCodeStandard)
	if afterSavedTest.tier.LastTestStatus != InvocationStatusSuccess ||
		afterSavedTest.tier.LastTestAt == nil ||
		sameOptionalTime(afterSavedTest.tier.LastTestAt, beforeSavedTest.tier.LastTestAt) {
		t.Fatalf("saved tier test did not update last test summary, before=%#v after=%#v", beforeSavedTest.tier, afterSavedTest.tier)
	}
}

// TestTierTestAppliesProviderCallTimeout verifies tier tests cancel provider
// calls through a service-level timeout even when the injected HTTP client has
// no timeout of its own.
func TestTierTestAppliesProviderCallTimeout(t *testing.T) {
	ctx := context.Background()
	prepareDatabase(t, ctx)
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}
	ctx = context.WithValue(ctx, providerBaseURLKey{}, "http://unit.local/v1")
	svc := New(testBizCtx{}, nil, client).(*serviceImpl)
	snapshot := snapshotTier(t, ctx, svc, TierCodeBasic)
	t.Cleanup(func() { restoreTier(t, ctx, snapshot) })

	providerID := createTestProvider(t, ctx, svc, "timeout")
	t.Cleanup(func() { deleteProviderFixture(t, ctx, providerID) })
	modelID := createTestModel(t, ctx, svc, providerID, "timeout-model")

	startedAt := time.Now()
	out, err := svc.TestTier(ctx, TierTestInput{
		Code:            TierCodeBasic,
		ProviderId:      providerID,
		ModelId:         modelID,
		ThinkingEffort:  string(aitext.ThinkingEffortLow),
		MaxOutputTokens: 128,
		Messages:        []aitext.Message{{Role: aitext.MessageRoleUser, Content: "ping"}},
		timeout:         20 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("test tier with provider timeout: %v", err)
	}
	if elapsed := time.Since(startedAt); elapsed > time.Second {
		t.Fatalf("expected provider test timeout to release quickly, elapsed=%s", elapsed)
	}
	if out.Status != InvocationStatusFailed || strings.TrimSpace(out.ErrorSummary) == "" {
		t.Fatalf("expected failed timeout output with summary, got %#v", out)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func sameOptionalTime(left *time.Time, right *time.Time) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.Equal(*right)
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
	firstBinding, err := reader.resolveTierBinding(ctx, CapabilityTypeText, CapabilityMethodGenerate, TierCodeAdvanced)
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
	secondBinding, err := reader.resolveTierBinding(ctx, CapabilityTypeText, CapabilityMethodGenerate, TierCodeAdvanced)
	if err != nil {
		t.Fatalf("resolve second binding: %v", err)
	}
	if secondBinding.ModelName != "revision-second-model" {
		t.Fatalf("expected peer cache invalidation to load second model, got %#v", secondBinding)
	}
}

// TestProviderSaveSynchronizesFormEndpointsTransactionally verifies provider
// metadata and fixed provider-form endpoints are committed or rolled back together.
func TestProviderSaveSynchronizesFormEndpointsTransactionally(t *testing.T) {
	ctx := context.Background()
	prepareDatabase(t, ctx)
	svc := New(testBizCtx{}, newMemoryCacheService(), nil).(*serviceImpl)

	providerID, err := svc.CreateProvider(ctx, ProviderSaveInput{
		Name:    "form-endpoints",
		Enabled: enabledYes,
		Endpoints: []ProviderEndpointSaveInput{
			{
				Protocol:     ProtocolOpenAI,
				BaseUrl:      "https://unit.example/openai/v1",
				SecretRef:    "sk-unit-openai",
				Enabled:      enabledYes,
				MetadataJson: "{}",
			},
			{
				Protocol:     ProtocolAnthropic,
				BaseUrl:      "https://unit.example/anthropic/v1",
				SecretRef:    "sk-unit-anthropic",
				Enabled:      enabledYes,
				MetadataJson: "{}",
			},
		},
	})
	if err != nil {
		t.Fatalf("create provider with endpoints: %v", err)
	}
	t.Cleanup(func() { deleteProviderFixture(t, ctx, providerID) })

	endpoints, err := svc.ListProviderEndpoints(ctx, ProviderEndpointListInput{ProviderId: providerID})
	if err != nil {
		t.Fatalf("list created endpoints: %v", err)
	}
	endpointsByProtocol := providerEndpointsByProtocol(endpoints)
	openaiEndpoint := endpointsByProtocol[ProtocolOpenAI]
	anthropicEndpoint := endpointsByProtocol[ProtocolAnthropic]
	if len(endpoints) != 2 || openaiEndpoint == nil || anthropicEndpoint == nil {
		t.Fatalf("expected two fixed endpoints, got %#v", endpoints)
	}
	if openaiEndpoint.SecretRef != "sk-**********ai" {
		t.Fatalf("expected masked OpenAI secret, got %#v", openaiEndpoint)
	}

	if err = svc.UpdateProvider(ctx, ProviderSaveInput{
		Id:      providerID,
		Name:    "form-endpoints-updated",
		Enabled: enabledYes,
		Endpoints: []ProviderEndpointSaveInput{
			{
				Id:           openaiEndpoint.Id,
				Protocol:     ProtocolOpenAI,
				BaseUrl:      "https://unit.example/openai/v2",
				SecretRef:    openaiEndpoint.SecretRef,
				Enabled:      enabledYes,
				MetadataJson: "{}",
			},
			{
				Id:       anthropicEndpoint.Id,
				Protocol: ProtocolAnthropic,
				BaseUrl:  "",
				Enabled:  enabledYes,
			},
		},
	}); err != nil {
		t.Fatalf("update provider endpoints: %v", err)
	}
	endpoints, err = svc.ListProviderEndpoints(ctx, ProviderEndpointListInput{ProviderId: providerID})
	if err != nil {
		t.Fatalf("list updated endpoints: %v", err)
	}
	endpointsByProtocol = providerEndpointsByProtocol(endpoints)
	openaiEndpoint = endpointsByProtocol[ProtocolOpenAI]
	if len(endpoints) != 1 || openaiEndpoint == nil || openaiEndpoint.BaseUrl != "https://unit.example/openai/v2" {
		t.Fatalf("expected one updated OpenAI endpoint, got %#v", endpoints)
	}

	if _, err = svc.CreateModel(ctx, ModelSaveInput{
		ProviderId:       providerID,
		EndpointId:       openaiEndpoint.Id,
		CapabilityType:   CapabilityTypeText,
		CapabilityMethod: CapabilityMethodGenerate,
		ModelName:        "form-endpoint-model",
		Protocol:         ProtocolOpenAI,
		Enabled:          enabledYes,
	}); err != nil {
		t.Fatalf("create model referencing endpoint: %v", err)
	}
	err = svc.UpdateProvider(ctx, ProviderSaveInput{
		Id:      providerID,
		Name:    "form-endpoints-rollback",
		Enabled: enabledNo,
		Endpoints: []ProviderEndpointSaveInput{{
			Id:       openaiEndpoint.Id,
			Protocol: ProtocolOpenAI,
			BaseUrl:  "",
			Enabled:  enabledYes,
		}},
	})
	if !bizerr.Is(err, CodeProviderEndpointInUse) {
		t.Fatalf("expected endpoint in-use rollback, got %v", err)
	}
	provider, err := svc.GetProvider(ctx, providerID)
	if err != nil {
		t.Fatalf("get provider after rollback: %v", err)
	}
	if provider.Name != "form-endpoints-updated" || provider.Enabled != enabledYes {
		t.Fatalf("expected provider metadata rollback, got %#v", provider)
	}
}

// TestMultimodalMetadataManagementFlow verifies endpoint CRUD, explicit model
// capability declarations, method defaults, method-scoped tier binding, and
// provider operation projection masking.
func TestMultimodalMetadataManagementFlow(t *testing.T) {
	ctx := context.Background()
	prepareDatabase(t, ctx)
	svc := New(testBizCtx{}, newMemoryCacheService(), nil).(*serviceImpl)

	providerID := createTestProvider(t, ctx, svc, "multimodal")
	t.Cleanup(func() { deleteProviderFixture(t, ctx, providerID) })
	snapshot := snapshotTierFor(t, ctx, svc, "image", "generate", TierCodeBasic)
	t.Cleanup(func() { restoreTier(t, ctx, snapshot) })

	endpointID, err := svc.CreateProviderEndpoint(ctx, ProviderEndpointSaveInput{
		ProviderId:   providerID,
		Protocol:     ProtocolOpenAICompatible,
		BaseUrl:      "https://unit.example/v1",
		SecretRef:    "sk-unit-endpoint",
		Enabled:      enabledYes,
		MetadataJson: `{"region":"unit"}`,
	})
	if err != nil {
		t.Fatalf("create endpoint: %v", err)
	}
	endpoints, err := svc.ListProviderEndpoints(ctx, ProviderEndpointListInput{
		ProviderId: providerID,
		Protocol:   ProtocolOpenAICompatible,
	})
	if err != nil {
		t.Fatalf("list endpoints: %v", err)
	}
	if len(endpoints) != 1 || endpoints[0].SecretRef != "sk-**********nt" || endpoints[0].BaseUrl != "https://unit.example/v1" {
		t.Fatalf("unexpected endpoint projection: %#v", endpoints)
	}
	if err = svc.UpdateProviderEndpoint(ctx, ProviderEndpointSaveInput{
		Id:           endpointID,
		ProviderId:   providerID,
		Protocol:     ProtocolOpenAICompatible,
		BaseUrl:      "https://unit.example/v2",
		SecretRef:    endpoints[0].SecretRef,
		Enabled:      enabledYes,
		MetadataJson: `{"region":"unit-2"}`,
	}); err != nil {
		t.Fatalf("update endpoint: %v", err)
	}

	modelID, err := svc.CreateModel(ctx, ModelSaveInput{
		ProviderId:       providerID,
		EndpointId:       endpointID,
		CapabilityType:   "image",
		CapabilityMethod: "generate",
		ModelName:        "unit-image-model",
		Protocol:         ProtocolOpenAICompatible,
		MaxInputTokens:   1024,
		MaxOutputTokens:  2048,
		Enabled:          enabledYes,
	})
	if err != nil {
		t.Fatalf("create image model: %v", err)
	}
	capabilities, err := svc.ListModelCapabilities(ctx, modelID)
	if err != nil {
		t.Fatalf("list model capabilities: %v", err)
	}
	if len(capabilities) != 1 || capabilities[0].CapabilityType != "image" || capabilities[0].CapabilityMethod != "generate" || capabilities[0].EndpointId != endpointID {
		t.Fatalf("expected generated image capability, got %#v", capabilities)
	}
	imageModels, err := svc.ListModels(ctx, ModelListInput{
		ProviderId:       providerID,
		PageNum:          1,
		PageSize:         10,
		CapabilityType:   "image",
		CapabilityMethod: "generate",
		Enabled:          intPtr(enabledYes),
	})
	if err != nil {
		t.Fatalf("list image models: %v", err)
	}
	if imageModels.Total != 1 ||
		imageModels.List[0].CapabilityType != "image" ||
		imageModels.List[0].MaxOutputTokens != 2048 {
		t.Fatalf("expected image model projection from capability table, got %#v", imageModels)
	}
	if err = svc.UpsertModelCapabilities(ctx, modelID, []ModelCapabilitySaveInput{{
		EndpointId:        endpointID,
		CapabilityType:    "image",
		CapabilityMethod:  "generate",
		InputModalities:   []string{"text"},
		OutputModalities:  []string{"image"},
		MaxInputTokens:    2048,
		MaxOutputTokens:   4096,
		MaxInputAssets:    1,
		MaxOutputAssets:   2,
		MaxAssetBytes:     8 * 1024 * 1024,
		SupportsOperation: enabledYes,
		DefaultParamsJson: `{"size":"1024x1024"}`,
		Enabled:           enabledYes,
		SupportsStreaming: enabledNo,
	}}); err != nil {
		t.Fatalf("upsert model capability: %v", err)
	}
	capabilities, err = svc.ListModelCapabilities(ctx, modelID)
	if err != nil {
		t.Fatalf("list updated model capabilities: %v", err)
	}
	if len(capabilities) != 1 || capabilities[0].MaxOutputAssets != 2 || capabilities[0].SupportsOperation != enabledYes {
		t.Fatalf("expected updated capability projection, got %#v", capabilities)
	}
	if err = svc.UpdateTier(ctx, TierUpdateInput{
		CapabilityType:   "image",
		CapabilityMethod: "generate",
		Code:             TierCodeBasic,
		ProviderId:       providerID,
		ModelId:          modelID,
		Enabled:          enabledYes,
	}); err != nil {
		t.Fatalf("bind image tier: %v", err)
	}
	if _, err = svc.resolveTierBinding(ctx, "image", "generate", TierCodeBasic); err != nil {
		t.Fatalf("resolve image tier binding: %v", err)
	}
	if err = svc.DeleteProviderEndpoint(ctx, providerID, endpointID); !bizerr.Is(err, CodeProviderEndpointInUse) {
		t.Fatalf("expected endpoint in-use protection, got %v", err)
	}

	if err = svc.UpdateMethodDefault(ctx, MethodDefaultParamSaveInput{
		CapabilityType:    "image",
		CapabilityMethod:  "generate",
		DefaultParamsJson: `{"count":2,"size":"1024x1024"}`,
		Enabled:           enabledYes,
	}); err != nil {
		t.Fatalf("update method default: %v", err)
	}
	defaults, err := svc.ListMethodDefaults(ctx)
	if err != nil {
		t.Fatalf("list method defaults: %v", err)
	}
	foundDefault := false
	for _, item := range defaults {
		if item.CapabilityType == "image" && item.CapabilityMethod == "generate" {
			foundDefault = item.DefaultParamsJson == `{"count":2,"size":"1024x1024"}`
			break
		}
	}
	if !foundDefault {
		t.Fatalf("expected updated image.generate defaults, got %#v", defaults)
	}

	expiresAt := time.Now().Add(time.Hour)
	if _, err = dao.ProviderOperation.Ctx(ctx).Data(do.ProviderOperation{
		OperationRef:     "op-unit-multimodal",
		CapabilityType:   "video",
		CapabilityMethod: "generate",
		Purpose:          "unit.multimodal",
		SourcePluginId:   "unit-plugin",
		ProviderId:       providerID,
		ModelId:          modelID,
		ProviderName:     "Unit Provider",
		ModelName:        "unit-image-model",
		Protocol:         ProtocolOpenAICompatible,
		Status:           "running",
		NextPollAfterMs:  3000,
		ExpiresAt:        &expiresAt,
		AssetSummaryJson: `{"assetRef":"asset://unit/video"}`,
		ErrorCode:        CodeProviderHTTPError.RuntimeCode(),
		ErrorSummary:     "provider returned sk-unit-secret in response",
	}).Insert(); err != nil {
		t.Fatalf("insert provider operation: %v", err)
	}
	operations, err := svc.ListProviderOperations(ctx, ProviderOperationListInput{
		PageNum:          1,
		PageSize:         10,
		CapabilityType:   "video",
		CapabilityMethod: "generate",
		Purpose:          "unit.multimodal",
		Status:           "running",
	})
	if err != nil {
		t.Fatalf("list provider operations: %v", err)
	}
	if operations.Total != 1 || len(operations.List) != 1 || strings.Contains(operations.List[0].ErrorSummary, "sk-unit-secret") {
		t.Fatalf("expected one masked provider operation, got %#v", operations)
	}
	operation, err := svc.GetProviderOperation(ctx, "op-unit-multimodal")
	if err != nil {
		t.Fatalf("get provider operation: %v", err)
	}
	if operation.OperationRef != "op-unit-multimodal" || strings.Contains(operation.ErrorSummary, "sk-unit-secret") {
		t.Fatalf("unexpected provider operation projection: %#v", operation)
	}
}

type testBizCtx struct{}

func (testBizCtx) Current(context.Context) plugincontract.CurrentContext {
	return plugincontract.CurrentContext{UserID: 1, TenantID: 0, PlatformBypass: true}
}

func providerEndpointsByProtocol(items []*ProviderEndpointItem) map[string]*ProviderEndpointItem {
	out := make(map[string]*ProviderEndpointItem, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out[item.Protocol] = item
	}
	return out
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
		sqlPaths, readErr := filepath.Glob(filepath.Clean("../../../../manifest/sql/*.sql"))
		if readErr != nil {
			installSQLError = readErr
			return
		}
		sort.Strings(sqlPaths)
		for _, sqlPath := range sqlPaths {
			content, readErr := os.ReadFile(sqlPath)
			if readErr != nil {
				installSQLError = readErr
				return
			}
			if strings.TrimSpace(string(content)) == "" {
				continue
			}
			if _, execErr := g.DB().Exec(ctx, string(content)); execErr != nil {
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
		Name:    "unit-provider-" + suffix + "-" + time.Now().Format("150405.000000000"),
		Enabled: enabledYes,
	})
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}
	if _, err = svc.CreateProviderEndpoint(ctx, ProviderEndpointSaveInput{
		ProviderId: id,
		Protocol:   ProtocolOpenAI,
		BaseUrl:    testProviderBaseURL(ctx),
		SecretRef:  "unit-secret",
		Enabled:    enabledYes,
	}); err != nil {
		t.Fatalf("create provider endpoint: %v", err)
	}
	return id
}

func createTestModel(t *testing.T, ctx context.Context, svc *serviceImpl, providerID int64, modelName string) int64 {
	t.Helper()
	id, err := svc.CreateModel(ctx, ModelSaveInput{
		ProviderId:       providerID,
		EndpointId:       defaultTestEndpointID(t, ctx, svc, providerID, ProtocolOpenAI),
		CapabilityType:   CapabilityTypeText,
		CapabilityMethod: CapabilityMethodGenerate,
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

func defaultTestEndpointID(t *testing.T, ctx context.Context, svc *serviceImpl, providerID int64, protocol string) int64 {
	t.Helper()
	endpoint, err := svc.enabledEndpointForProtocol(ctx, providerID, protocol)
	if err != nil {
		t.Fatalf("load default test endpoint: %v", err)
	}
	return endpoint.Id
}

func intPtr(value int) *int {
	return &value
}

func deleteProviderFixture(t *testing.T, ctx context.Context, providerID int64) {
	t.Helper()
	statements := []string{
		`DELETE FROM plugin_linapro_ai_invocation WHERE provider_id = $1`,
		`DELETE FROM plugin_linapro_ai_provider_operation WHERE provider_id = $1`,
		`DELETE FROM plugin_linapro_ai_tier_binding WHERE provider_id = $1`,
		`DELETE FROM plugin_linapro_ai_model_capability WHERE model_id IN (SELECT id FROM plugin_linapro_ai_model WHERE provider_id = $1)`,
		`DELETE FROM plugin_linapro_ai_model WHERE provider_id = $1`,
		`DELETE FROM plugin_linapro_ai_provider_endpoint WHERE provider_id = $1`,
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
	return snapshotTierFor(t, ctx, svc, CapabilityTypeText, CapabilityMethodGenerate, code)
}

func snapshotTierFor(t *testing.T, ctx context.Context, svc *serviceImpl, capabilityType string, capabilityMethod string, code string) tierSnapshot {
	t.Helper()
	tier, err := svc.getTier(ctx, capabilityType, capabilityMethod, code)
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
