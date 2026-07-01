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

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"

	"lina-core/pkg/bizerr"
	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/plugin/capability/aicap/aitext"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/cachecap"
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
	snapshot := snapshotTier(t, ctx, svc, tierCodeBasic)
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
		Code:          tierCodeBasic,
		ProviderId:    providerID,
		ModelId:       modelID,
		DefaultEffort: string(aitext.ThinkingEffortLow),
		Enabled:       enabledYes,
	}); err != nil {
		t.Fatalf("update tier: %v", err)
	}
	if err = svc.UpdateTier(ctx, TierUpdateInput{
		Code:          tierCodeBasic,
		DefaultEffort: string(aitext.ThinkingEffortLow),
		Enabled:       enabledNo,
	}); err != nil {
		t.Fatalf("disable tier without rebinding: %v", err)
	}
	preserved := snapshotTier(t, ctx, svc, tierCodeBasic)
	if preserved.binding == nil || preserved.binding.ProviderId != providerID || preserved.binding.ModelId != modelID {
		t.Fatalf("expected disabling tier to preserve binding, got %#v", preserved.binding)
	}
	if err = svc.UpdateTier(ctx, TierUpdateInput{
		Code:          tierCodeBasic,
		ProviderId:    providerID,
		ModelId:       modelID,
		DefaultEffort: string(aitext.ThinkingEffortLow),
		Enabled:       enabledYes,
	}); err != nil {
		t.Fatalf("re-enable tier: %v", err)
	}
	syncOut, err := svc.SyncModels(ctx, ModelSyncInput{ProviderId: providerID, Protocol: protocolOpenAI})
	if err != nil {
		t.Fatalf("sync models: %v", err)
	}
	if syncOut.Created != 1 || syncOut.Kept != 1 {
		t.Fatalf("expected sync to batch-detect existing models, got %#v", syncOut)
	}
	if err = svc.UpdateModel(ctx, ModelSaveInput{
		Id:         modelID,
		EndpointId: defaultTestEndpointID(t, ctx, svc, providerID, protocolOpenAI),
		ModelName:  "flow-model-renamed",
		Protocol:   protocolOpenAI,
		Enabled:    enabledYes,
	}); err != nil {
		t.Fatalf("update model identity: %v", err)
	}
	capabilities, err := svc.ListModelCapabilities(ctx, modelID)
	if err != nil {
		t.Fatalf("list capabilities after model update: %v", err)
	}
	if len(capabilities) != 0 {
		t.Fatalf("model update must not create capability declarations, got %#v", capabilities)
	}
	models, err := svc.ListModels(ctx, ModelListInput{
		ProviderId: providerID,
		PageNum:    1,
		PageSize:   10,
		Enabled:    intPtr(enabledYes),
	})
	if err != nil {
		t.Fatalf("list models after model update: %v", err)
	}
	foundRenamed := false
	for _, model := range models.List {
		if model != nil && model.ModelName == "flow-model-renamed" {
			foundRenamed = true
			break
		}
	}
	if !foundRenamed {
		t.Fatalf("expected model list projection from identity table, got %#v", models)
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
		CapabilityMethod: capabilityMethodGenerate,
		Purpose:          "unit.flow",
		Status:           invocationStatusSuccess,
	})
	if err != nil {
		t.Fatalf("list invocations: %v", err)
	}
	if logs.Total < 1 || logs.List[0].CapabilityMethod != capabilityMethodGenerate || strings.Contains(logs.List[0].ErrorSummary, "full prompt") {
		t.Fatalf("expected masked invocation log, got %#v", logs)
	}
}

// TestProviderModelSummaryAggregatesByNameAndDeleteRemovesGroup verifies the
// provider list hides protocol-level duplicates and delete removes the whole
// provider-local model-name group.
func TestProviderModelSummaryAggregatesByNameAndDeleteRemovesGroup(t *testing.T) {
	ctx := context.Background()
	prepareDatabase(t, ctx)
	svc := New(testBizCtx{}, newMemoryCacheService(), nil).(*serviceImpl)

	providerID := createTestProvider(t, ctx, svc, "aggregate-model-name")
	t.Cleanup(func() { deleteProviderFixture(t, ctx, providerID) })
	anthropicEndpointID, err := svc.CreateProviderEndpoint(ctx, ProviderEndpointSaveInput{
		ProviderId: providerID,
		Protocol:   protocolAnthropic,
		BaseUrl:    "https://unit.example/anthropic/v1",
		SecretRef:  "unit-secret",
		Enabled:    enabledYes,
	})
	if err != nil {
		t.Fatalf("create anthropic endpoint: %v", err)
	}

	sharedModelName := "aggregate-delete-model"
	openAIModelID := createTestModel(t, ctx, svc, providerID, sharedModelName)
	if _, err = svc.CreateModel(ctx, ModelSaveInput{
		ProviderId: providerID,
		EndpointId: anthropicEndpointID,
		ModelName:  sharedModelName,
		Protocol:   protocolAnthropic,
		Enabled:    enabledYes,
	}); err != nil {
		t.Fatalf("create duplicate-name anthropic model: %v", err)
	}

	provider, err := svc.GetProvider(ctx, providerID)
	if err != nil {
		t.Fatalf("get provider with duplicate-name models: %v", err)
	}
	summaryCount := 0
	for _, summary := range provider.Models {
		if summary != nil && summary.ModelName == sharedModelName {
			summaryCount++
		}
	}
	if summaryCount != 1 {
		t.Fatalf("expected provider model summaries to aggregate by model name, got %#v", provider.Models)
	}

	if err = svc.DeleteModel(ctx, openAIModelID); err != nil {
		t.Fatalf("delete duplicate-name model group: %v", err)
	}
	models, err := svc.ListModels(ctx, ModelListInput{
		ProviderId: providerID,
		PageNum:    1,
		PageSize:   100,
	})
	if err != nil {
		t.Fatalf("list models after grouped delete: %v", err)
	}
	for _, model := range models.List {
		if model != nil && model.ModelName == sharedModelName {
			t.Fatalf("expected grouped delete to remove all same-name rows, got %#v", models.List)
		}
	}

	referencedModelName := "aggregate-delete-referenced"
	referencedOpenAIID := createTestModel(t, ctx, svc, providerID, referencedModelName)
	referencedAnthropicID, err := svc.CreateModel(ctx, ModelSaveInput{
		ProviderId: providerID,
		EndpointId: anthropicEndpointID,
		ModelName:  referencedModelName,
		Protocol:   protocolAnthropic,
		Enabled:    enabledYes,
	})
	if err != nil {
		t.Fatalf("create referenced duplicate-name model: %v", err)
	}
	snapshot := snapshotTier(t, ctx, svc, tierCodeBasic)
	t.Cleanup(func() { restoreTier(t, ctx, snapshot) })
	if err = svc.UpdateTier(ctx, TierUpdateInput{
		Code:          tierCodeBasic,
		ProviderId:    providerID,
		ModelId:       referencedAnthropicID,
		DefaultEffort: string(aitext.ThinkingEffortLow),
		Enabled:       enabledYes,
	}); err != nil {
		t.Fatalf("bind referenced duplicate-name model: %v", err)
	}
	if err = svc.DeleteModel(ctx, referencedOpenAIID); !bizerr.Is(err, CodeModelInUse) {
		t.Fatalf("expected grouped delete to reject referenced peer model, got %v", err)
	}
	models, err = svc.ListModels(ctx, ModelListInput{
		ProviderId: providerID,
		PageNum:    1,
		PageSize:   100,
	})
	if err != nil {
		t.Fatalf("list models after rejected grouped delete: %v", err)
	}
	remaining := 0
	for _, model := range models.List {
		if model != nil && model.ModelName == referencedModelName {
			remaining++
		}
	}
	if remaining != 2 {
		t.Fatalf("expected rejected grouped delete to preserve both same-name rows, got %#v", models.List)
	}
}

// TestCleanInvocationsDeletesAllOrTimeRange verifies range cleanup and full
// cleanup both execute in the database and report affected row counts.
func TestCleanInvocationsDeletesAllOrTimeRange(t *testing.T) {
	ctx := context.Background()
	prepareDatabase(t, ctx)
	svc := New(testBizCtx{}, nil, nil).(*serviceImpl)
	const rollbackMessage = "rollback invocation cleanup test transaction"
	err := dao.Invocation.Transaction(ctx, func(txCtx context.Context, _ gdb.TX) error {
		var (
			base         = time.Now().Add(-2 * time.Hour).Truncate(time.Millisecond)
			prefix       = base.Format("20060102150405")
			oldRequestID = "unit-clean-old-" + prefix
			midRequestID = "unit-clean-mid-" + prefix
			newRequestID = "unit-clean-new-" + prefix
		)
		insertInvocationFixture(t, txCtx, oldRequestID, base)
		insertInvocationFixture(t, txCtx, midRequestID, base.Add(time.Hour))
		insertInvocationFixture(t, txCtx, newRequestID, base.Add(2*time.Hour))

		deleted, err := svc.CleanInvocations(txCtx, InvocationCleanInput{
			StartedAt: base.Add(30 * time.Minute).UnixMilli(),
			EndedAt:   base.Add(90 * time.Minute).UnixMilli(),
		})
		if err != nil {
			t.Fatalf("range clean invocations: %v", err)
		}
		if deleted != 1 || invocationExists(t, txCtx, midRequestID) {
			t.Fatalf("expected range cleanup to delete only middle log, deleted=%d", deleted)
		}
		if !invocationExists(t, txCtx, oldRequestID) || !invocationExists(t, txCtx, newRequestID) {
			t.Fatalf("range cleanup deleted records outside time bounds")
		}

		deleted, err = svc.CleanInvocations(txCtx, InvocationCleanInput{})
		if err != nil {
			t.Fatalf("full clean invocations: %v", err)
		}
		if deleted < 2 ||
			invocationExists(t, txCtx, oldRequestID) ||
			invocationExists(t, txCtx, newRequestID) {
			t.Fatalf("expected full cleanup to delete remaining fixture logs, deleted=%d", deleted)
		}
		return gerror.New(rollbackMessage)
	})
	if err == nil || err.Error() != rollbackMessage {
		t.Fatalf("expected transaction rollback marker %q, got %v", rollbackMessage, err)
	}
}

// TestCleanupExpiredInvocationsDeletesOnlyOldLogs verifies retention cleanup
// removes logs older than the configured global retention boundary.
func TestCleanupExpiredInvocationsDeletesOnlyOldLogs(t *testing.T) {
	ctx := context.Background()
	prepareDatabase(t, ctx)
	svc := New(testBizCtx{}, nil, nil).(*serviceImpl)
	const rollbackMessage = "rollback invocation retention cleanup test transaction"
	err := dao.Invocation.Transaction(ctx, func(txCtx context.Context, _ gdb.TX) error {
		var (
			base         = time.Now().AddDate(0, 0, -2).Truncate(time.Millisecond)
			prefix       = time.Now().Format("20060102150405")
			oldRequestID = "unit-retention-old-" + prefix
			newRequestID = "unit-retention-new-" + prefix
		)
		insertInvocationFixture(t, txCtx, oldRequestID, base)
		insertInvocationFixture(t, txCtx, newRequestID, time.Now())

		deleted, err := svc.CleanupExpiredInvocations(txCtx, 1)
		if err != nil {
			t.Fatalf("retention clean invocations: %v", err)
		}
		if deleted < 1 {
			t.Fatalf("expected at least the old fixture log to be deleted, got %d", deleted)
		}
		if invocationExists(t, txCtx, oldRequestID) {
			t.Fatalf("expected old invocation log %s to be deleted", oldRequestID)
		}
		if !invocationExists(t, txCtx, newRequestID) {
			t.Fatalf("expected new invocation log %s to remain", newRequestID)
		}
		return gerror.New(rollbackMessage)
	})
	if err == nil || err.Error() != rollbackMessage {
		t.Fatalf("expected transaction rollback marker %q, got %v", rollbackMessage, err)
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
		Protocol:   protocolAnthropic,
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
	existingModels, err := svc.ListAllModels(ctx, ModelGlobalListInput{
		Keyword:  "sync-existing",
		PageNum:  1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("list kept model: %v", err)
	}
	if existingModels.Total != 1 || existingModels.List[0].ModelName != "sync-existing" {
		t.Fatalf("expected kept model identity to remain, got %#v", existingModels)
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
		models.List[0].Protocol != protocolOpenAI {
		t.Fatalf("expected synced model provider and endpoint projection, got %#v", models.List[0])
	}
	capabilities, err := svc.ListModelCapabilities(ctx, models.List[0].Id)
	if err != nil {
		t.Fatalf("list synced model capabilities: %v", err)
	}
	if len(capabilities) != 0 {
		t.Fatalf("sync must not infer text.generate capability, got %#v", capabilities)
	}
	filteredModels, err := svc.ListModels(ctx, ModelListInput{
		ProviderId: providerID,
		PageNum:    1,
		PageSize:   10,
		Enabled:    intPtr(enabledYes),
	})
	if err != nil {
		t.Fatalf("list text models after sync: %v", err)
	}
	if filteredModels.Total != 2 {
		t.Fatalf("expected capability filters to be removed from provider model list, got %#v", filteredModels)
	}
	unfilteredModels, err := svc.ListModels(ctx, ModelListInput{
		ProviderId: providerID,
		PageNum:    1,
		PageSize:   10,
		Enabled:    intPtr(enabledYes),
	})
	if err != nil {
		t.Fatalf("list unfiltered provider models after sync: %v", err)
	}
	if unfilteredModels.Total != 2 {
		t.Fatalf("expected unfiltered provider model list to include unclassified models, got %#v", unfilteredModels)
	}
	provider, err := svc.GetProvider(ctx, providerID)
	if err != nil {
		t.Fatalf("get provider after sync: %v", err)
	}
	if provider.ModelCount != 2 || len(provider.Models) != 2 {
		t.Fatalf("expected provider summary to include classified and unclassified models, got %#v", provider)
	}

	failingProviderID, err := svc.CreateProvider(ctx, ProviderSaveInput{
		Name:    "unit-provider-all-failing-sync-" + time.Now().Format("150405.000000000"),
		Enabled: enabledYes,
	})
	if err != nil {
		t.Fatalf("create all-failing provider: %v", err)
	}
	t.Cleanup(func() { deleteProviderFixture(t, ctx, failingProviderID) })
	for _, protocol := range []string{protocolOpenAI, protocolAnthropic} {
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

// TestConfirmedRemoteCapabilitiesInsertMissingOnly verifies confirmed remote
// capabilities can be added without overwriting existing manual declarations.
func TestConfirmedRemoteCapabilitiesInsertMissingOnly(t *testing.T) {
	ctx := context.Background()
	prepareDatabase(t, ctx)
	server := testOpenAIServer(t)
	ctx = context.WithValue(ctx, providerBaseURLKey{}, server.URL+"/v1")
	svc := New(testBizCtx{}, nil, server.Client()).(*serviceImpl)

	providerID := createTestProvider(t, ctx, svc, "confirmed-capability")
	t.Cleanup(func() { deleteProviderFixture(t, ctx, providerID) })
	modelID := createTestModel(t, ctx, svc, providerID, "confirmed-capability-model")
	endpointID := defaultTestEndpointID(t, ctx, svc, providerID, protocolOpenAI)
	if err := svc.UpsertModelCapabilities(ctx, modelID, []ModelCapabilitySaveInput{{
		EndpointId:        endpointID,
		CapabilityType:    capabilityTypeText,
		CapabilityMethod:  capabilityMethodGenerate,
		InputModalities:   []string{"manual"},
		OutputModalities:  []string{"text"},
		MaxOutputTokens:   256,
		SupportsOperation: enabledNo,
		Enabled:           enabledYes,
	}}); err != nil {
		t.Fatalf("create manual text state: %v", err)
	}
	model := &entity.Model{
		Id:         modelID,
		ProviderId: providerID,
		EndpointId: endpointID,
		Protocol:   protocolOpenAI,
	}

	existingCapabilityKeys, err := svc.modelCapabilityKeysByModel(ctx, []int64{modelID})
	if err != nil {
		t.Fatalf("load existing capability keys: %v", err)
	}
	if err := svc.insertConfirmedRemoteModelCapabilities(ctx, existingCapabilityKeys, model, endpointID, []remoteModelCapability{
		{
			CapabilityType:   capabilityTypeText,
			CapabilityMethod: capabilityMethodGenerate,
			InputModalities:  []string{"remote"},
			OutputModalities: []string{"remote"},
		},
		{
			CapabilityType:   "image",
			CapabilityMethod: "generate",
			InputModalities:  []string{"text"},
			OutputModalities: []string{"image"},
		},
	}); err != nil {
		t.Fatalf("insert confirmed remote capabilities: %v", err)
	}

	capabilities, err := svc.ListModelCapabilities(ctx, modelID)
	if err != nil {
		t.Fatalf("list confirmed capabilities: %v", err)
	}
	if len(capabilities) != 2 {
		t.Fatalf("expected manual text capability plus inserted image capability, got %#v", capabilities)
	}
	for _, capability := range capabilities {
		if capability.CapabilityType == capabilityTypeText &&
			(capability.MaxOutputTokens != 256 || strings.Join(capability.InputModalities, ",") == "remote") {
			t.Fatalf("remote sync must not overwrite manual text capability, got %#v", capability)
		}
	}
}

// TestModelCapabilityReplacementProtectsTierBindings verifies the legacy
// capability metadata API can replace explicit declarations, while tier-bound
// declarations remain protected even though model management no longer uses
// them as the candidate source.
func TestModelCapabilityReplacementAllowsMethodEditAndProtectsTierBindings(t *testing.T) {
	ctx := context.Background()
	prepareDatabase(t, ctx)
	svc := New(testBizCtx{}, nil, nil).(*serviceImpl)

	providerID := createTestProvider(t, ctx, svc, "capability-replace")
	t.Cleanup(func() { deleteProviderFixture(t, ctx, providerID) })
	endpointID := defaultTestEndpointID(t, ctx, svc, providerID, protocolOpenAI)
	modelID := createTestModel(t, ctx, svc, providerID, "capability-replace-model")

	if err := svc.UpsertModelCapabilities(ctx, modelID, []ModelCapabilitySaveInput{{
		EndpointId:        endpointID,
		CapabilityType:    "document",
		CapabilityMethod:  "analyze",
		InputModalities:   []string{"document"},
		OutputModalities:  []string{"text"},
		MaxInputAssets:    1,
		MaxOutputTokens:   333,
		SupportsOperation: enabledNo,
		Enabled:           enabledYes,
	}}); err != nil {
		t.Fatalf("replace model capability method: %v", err)
	}
	capabilities, err := svc.ListModelCapabilities(ctx, modelID)
	if err != nil {
		t.Fatalf("list replaced model capabilities: %v", err)
	}
	if len(capabilities) != 1 ||
		capabilities[0].CapabilityType != "document" ||
		capabilities[0].CapabilityMethod != "analyze" ||
		capabilities[0].MaxOutputTokens != 333 {
		t.Fatalf("expected replacement capability only, got %#v", capabilities)
	}
	documentModels, err := svc.ListModels(ctx, ModelListInput{
		ProviderId: providerID,
		PageNum:    1,
		PageSize:   10,
	})
	if err != nil {
		t.Fatalf("list provider models after replacement: %v", err)
	}
	if documentModels.Total != 1 ||
		documentModels.List[0].ModelName != "capability-replace-model" {
		t.Fatalf("expected model list to ignore capability declarations, got %#v", documentModels)
	}

	referencedModelID := createTestModel(t, ctx, svc, providerID, "capability-replace-referenced")
	if err := svc.UpsertModelCapabilities(ctx, referencedModelID, []ModelCapabilitySaveInput{{
		EndpointId:        endpointID,
		CapabilityType:    capabilityTypeText,
		CapabilityMethod:  capabilityMethodGenerate,
		InputModalities:   []string{"text"},
		OutputModalities:  []string{"text"},
		SupportsOperation: enabledNo,
		Enabled:           enabledYes,
	}}); err != nil {
		t.Fatalf("create referenced text state: %v", err)
	}
	snapshot := snapshotTier(t, ctx, svc, tierCodeStandard)
	t.Cleanup(func() { restoreTier(t, ctx, snapshot) })
	if err = svc.UpdateTier(ctx, TierUpdateInput{
		Code:       tierCodeStandard,
		ProviderId: providerID,
		ModelId:    referencedModelID,
		Enabled:    enabledYes,
	}); err != nil {
		t.Fatalf("bind referenced model state: %v", err)
	}
	err = svc.UpsertModelCapabilities(ctx, referencedModelID, []ModelCapabilitySaveInput{{
		EndpointId:       endpointID,
		CapabilityType:   "image",
		CapabilityMethod: "generate",
		Enabled:          enabledYes,
	}})
	if !bizerr.Is(err, CodeModelInUse) {
		t.Fatalf("expected tier-bound capability replacement to be rejected, got %v", err)
	}
	capabilities, err = svc.ListModelCapabilities(ctx, referencedModelID)
	if err != nil {
		t.Fatalf("list referenced model capabilities after rejected replacement: %v", err)
	}
	if len(capabilities) != 1 ||
		capabilities[0].CapabilityType != capabilityTypeText ||
		capabilities[0].CapabilityMethod != capabilityMethodGenerate {
		t.Fatalf("expected referenced capability to remain unchanged, got %#v", capabilities)
	}
}

// TestTierBindingAllowsModelWithoutCapabilityDeclaration verifies text.generate
// tiers bind enabled models without requiring model-side capability declarations.
func TestTierBindingAllowsModelWithoutCapabilityDeclaration(t *testing.T) {
	ctx := context.Background()
	prepareDatabase(t, ctx)
	server := testOpenAIServer(t)
	ctx = context.WithValue(ctx, providerBaseURLKey{}, server.URL+"/v1")
	svc := New(testBizCtx{}, nil, server.Client()).(*serviceImpl)
	snapshot := snapshotTier(t, ctx, svc, tierCodeBasic)
	t.Cleanup(func() { restoreTier(t, ctx, snapshot) })

	providerID := createTestProvider(t, ctx, svc, "wrong-method")
	t.Cleanup(func() { deleteProviderFixture(t, ctx, providerID) })
	modelID, err := svc.CreateModel(ctx, ModelSaveInput{
		ProviderId: providerID,
		EndpointId: defaultTestEndpointID(t, ctx, svc, providerID, protocolOpenAI),
		ModelName:  "capability-free-model",
		Protocol:   protocolOpenAI,
		Enabled:    enabledYes,
	})
	if err != nil {
		t.Fatalf("create capability-free model: %v", err)
	}

	err = svc.UpdateTier(ctx, TierUpdateInput{
		CapabilityType:   capabilityTypeText,
		CapabilityMethod: capabilityMethodGenerate,
		Code:             tierCodeBasic,
		ProviderId:       providerID,
		ModelId:          modelID,
		DefaultEffort:    string(aitext.ThinkingEffortMax),
		Enabled:          enabledYes,
	})
	if err != nil {
		t.Fatalf("expected tier binding to allow model without declaration, got %v", err)
	}
	binding, err := svc.resolveTierBinding(ctx, capabilityTypeText, capabilityMethodGenerate, tierCodeBasic)
	if err != nil {
		t.Fatalf("resolve capability-free tier binding: %v", err)
	}
	if binding.ModelName != "capability-free-model" {
		t.Fatalf("expected capability-free binding, got %#v", binding)
	}
	if binding.DefaultEffort != string(aitext.ThinkingEffortMax) {
		t.Fatalf("expected max effort to be accepted without model declaration, got %#v", binding)
	}
}

// TestMultimodalTierBindingRequiresModelCapabilityDeclaration verifies non-text
// tiers cannot bind an unclassified or wrong-method model.
func TestMultimodalTierBindingRequiresModelCapabilityDeclaration(t *testing.T) {
	ctx := context.Background()
	prepareDatabase(t, ctx)
	server := testOpenAIServer(t)
	ctx = context.WithValue(ctx, providerBaseURLKey{}, server.URL+"/v1")
	svc := New(testBizCtx{}, nil, server.Client()).(*serviceImpl)
	snapshot := snapshotTierFor(t, ctx, svc, "document", "analyze", tierCodeStandard)
	t.Cleanup(func() { restoreTier(t, ctx, snapshot) })

	providerID := createTestProvider(t, ctx, svc, "document-capability-required")
	t.Cleanup(func() { deleteProviderFixture(t, ctx, providerID) })
	modelID, err := svc.CreateModel(ctx, ModelSaveInput{
		ProviderId: providerID,
		EndpointId: defaultTestEndpointID(t, ctx, svc, providerID, protocolOpenAI),
		ModelName:  "capability-free-document-model",
		Protocol:   protocolOpenAI,
		Enabled:    enabledYes,
	})
	if err != nil {
		t.Fatalf("create capability-free model: %v", err)
	}

	err = svc.UpdateTier(ctx, TierUpdateInput{
		CapabilityType:   "document",
		CapabilityMethod: "analyze",
		Code:             tierCodeStandard,
		ProviderId:       providerID,
		ModelId:          modelID,
		Enabled:          enabledYes,
	})
	if !bizerr.Is(err, CodeModelNotFound) {
		t.Fatalf("expected document tier binding to reject undeclared model capability, got %v", err)
	}

	_, err = svc.TestTier(ctx, TierTestInput{
		CapabilityType:   "document",
		CapabilityMethod: "analyze",
		Code:             tierCodeStandard,
		ProviderId:       providerID,
		ModelId:          modelID,
	})
	if !bizerr.Is(err, CodeModelNotFound) {
		t.Fatalf("expected document tier draft test to reject undeclared model capability, got %v", err)
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
	snapshot := snapshotTier(t, ctx, svc, tierCodeStandard)
	t.Cleanup(func() { restoreTier(t, ctx, snapshot) })

	providerID := createTestProvider(t, ctx, svc, "draft")
	t.Cleanup(func() { deleteProviderFixture(t, ctx, providerID) })
	modelID := createTestModel(t, ctx, svc, providerID, "draft-model")

	out, err := svc.TestTier(ctx, TierTestInput{
		Code:           tierCodeStandard,
		ProviderId:     providerID,
		ModelId:        modelID,
		ThinkingEffort: string(aitext.ThinkingEffortLow),
		Messages:       []aitext.Message{{Role: aitext.MessageRoleUser, Content: "ping"}},
	})
	if err != nil {
		t.Fatalf("test draft tier: %v", err)
	}
	if out.Status != invocationStatusSuccess || out.ModelName != "draft-model" {
		t.Fatalf("unexpected draft test output: %#v", out)
	}
	draftLogs, err := svc.ListInvocations(ctx, InvocationListInput{
		PageNum:          1,
		PageSize:         10,
		CapabilityType:   capabilityTypeText,
		CapabilityMethod: capabilityMethodGenerate,
		Purpose:          invocationPurposeTierTest,
		TierCode:         tierCodeStandard,
		SourcePluginId:   aitext.ProviderPluginID,
		ProviderId:       providerID,
		ModelId:          modelID,
		Status:           invocationStatusSuccess,
	})
	if err != nil {
		t.Fatalf("list draft tier test invocation logs: %v", err)
	}
	if draftLogs.Total != 1 ||
		draftLogs.List[0].ProviderName == "" ||
		draftLogs.List[0].ModelName != "draft-model" ||
		draftLogs.List[0].ThinkingEffort != string(aitext.ThinkingEffortLow) {
		t.Fatalf("expected draft tier test invocation log, got %#v", draftLogs)
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
	afterDraft := snapshotTier(t, ctx, svc, tierCodeStandard)
	if afterDraft.tier.LastTestStatus != snapshot.tier.LastTestStatus ||
		afterDraft.tier.LastTestLatencyMs != snapshot.tier.LastTestLatencyMs ||
		afterDraft.tier.LastTestErrorSummary != snapshot.tier.LastTestErrorSummary ||
		!sameOptionalTime(afterDraft.tier.LastTestAt, snapshot.tier.LastTestAt) {
		t.Fatalf("draft tier test changed last test summary, before=%#v after=%#v", snapshot.tier, afterDraft.tier)
	}

	if err = svc.UpdateTier(ctx, TierUpdateInput{
		Code:          tierCodeStandard,
		ProviderId:    providerID,
		ModelId:       modelID,
		DefaultEffort: string(aitext.ThinkingEffortLow),
		Enabled:       enabledYes,
	}); err != nil {
		t.Fatalf("save tier binding: %v", err)
	}
	beforeSavedTest := snapshotTier(t, ctx, svc, tierCodeStandard)
	savedOut, err := svc.TestTier(ctx, TierTestInput{
		Code:     tierCodeStandard,
		Messages: []aitext.Message{{Role: aitext.MessageRoleUser, Content: "ping"}},
	})
	if err != nil {
		t.Fatalf("test saved tier: %v", err)
	}
	if savedOut.Status != invocationStatusSuccess || savedOut.ModelName != "draft-model" {
		t.Fatalf("unexpected saved test output: %#v", savedOut)
	}
	afterSavedTest := snapshotTier(t, ctx, svc, tierCodeStandard)
	if afterSavedTest.tier.LastTestStatus != invocationStatusSuccess ||
		afterSavedTest.tier.LastTestAt == nil ||
		sameOptionalTime(afterSavedTest.tier.LastTestAt, beforeSavedTest.tier.LastTestAt) {
		t.Fatalf("saved tier test did not update last test summary, before=%#v after=%#v", beforeSavedTest.tier, afterSavedTest.tier)
	}
	savedLogs, err := svc.ListInvocations(ctx, InvocationListInput{
		PageNum:          1,
		PageSize:         10,
		CapabilityType:   capabilityTypeText,
		CapabilityMethod: capabilityMethodGenerate,
		Purpose:          invocationPurposeTierTest,
		TierCode:         tierCodeStandard,
		SourcePluginId:   aitext.ProviderPluginID,
		ProviderId:       providerID,
		ModelId:          modelID,
		Status:           invocationStatusSuccess,
	})
	if err != nil {
		t.Fatalf("list saved tier test invocation logs: %v", err)
	}
	if savedLogs.Total != 2 {
		t.Fatalf("expected draft and saved tier test invocation logs, got %#v", savedLogs)
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
	snapshot := snapshotTier(t, ctx, svc, tierCodeBasic)
	t.Cleanup(func() { restoreTier(t, ctx, snapshot) })

	providerID := createTestProvider(t, ctx, svc, "timeout")
	t.Cleanup(func() { deleteProviderFixture(t, ctx, providerID) })
	modelID := createTestModel(t, ctx, svc, providerID, "timeout-model")

	startedAt := time.Now()
	out, err := svc.TestTier(ctx, TierTestInput{
		Code:            tierCodeBasic,
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
	if out.Status != invocationStatusFailed || strings.TrimSpace(out.ErrorSummary) == "" {
		t.Fatalf("expected failed timeout output with summary, got %#v", out)
	}
	logs, err := svc.ListInvocations(ctx, InvocationListInput{
		PageNum:          1,
		PageSize:         10,
		CapabilityType:   capabilityTypeText,
		CapabilityMethod: capabilityMethodGenerate,
		Purpose:          invocationPurposeTierTest,
		TierCode:         tierCodeBasic,
		SourcePluginId:   aitext.ProviderPluginID,
		ProviderId:       providerID,
		ModelId:          modelID,
		Status:           invocationStatusFailed,
	})
	if err != nil {
		t.Fatalf("list failed tier test invocation logs: %v", err)
	}
	if logs.Total != 1 || strings.TrimSpace(logs.List[0].ErrorSummary) == "" {
		t.Fatalf("expected failed tier test invocation log, got %#v", logs)
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
	var (
		cacheSvc = newMemoryCacheService()
		writer   = New(testBizCtx{}, cacheSvc, server.Client()).(*serviceImpl)
		reader   = New(testBizCtx{}, cacheSvc, server.Client()).(*serviceImpl)
		snapshot = snapshotTier(t, ctx, writer, tierCodeAdvanced)
	)
	t.Cleanup(func() { restoreTier(t, ctx, snapshot) })

	firstProviderID := createTestProvider(t, ctx, writer, "revision-first")
	t.Cleanup(func() { deleteProviderFixture(t, ctx, firstProviderID) })
	firstModelID := createTestModel(t, ctx, writer, firstProviderID, "revision-first-model")
	if err := writer.UpdateTier(ctx, TierUpdateInput{
		Code:          tierCodeAdvanced,
		ProviderId:    firstProviderID,
		ModelId:       firstModelID,
		DefaultEffort: string(aitext.ThinkingEffortLow),
		Enabled:       enabledYes,
	}); err != nil {
		t.Fatalf("bind first tier: %v", err)
	}
	assertTierRevisionExpires(t, ctx, cacheSvc)
	firstBinding, err := reader.resolveTierBinding(ctx, capabilityTypeText, capabilityMethodGenerate, tierCodeAdvanced)
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
		Code:          tierCodeAdvanced,
		ProviderId:    secondProviderID,
		ModelId:       secondModelID,
		DefaultEffort: string(aitext.ThinkingEffortLow),
		Enabled:       enabledYes,
	}); err != nil {
		t.Fatalf("bind second tier: %v", err)
	}
	assertTierRevisionExpires(t, ctx, cacheSvc)
	secondBinding, err := reader.resolveTierBinding(ctx, capabilityTypeText, capabilityMethodGenerate, tierCodeAdvanced)
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
				Protocol:     protocolOpenAI,
				BaseUrl:      "https://unit.example/openai/v1",
				SecretRef:    "sk-unit-openai",
				Enabled:      enabledYes,
				MetadataJson: "{}",
			},
			{
				Protocol:     protocolAnthropic,
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
	var (
		endpointsByProtocol = providerEndpointsByProtocol(endpoints)
		openaiEndpoint      = endpointsByProtocol[protocolOpenAI]
		anthropicEndpoint   = endpointsByProtocol[protocolAnthropic]
	)
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
				Protocol:     protocolOpenAI,
				BaseUrl:      "https://unit.example/openai/v2",
				SecretRef:    openaiEndpoint.SecretRef,
				Enabled:      enabledYes,
				MetadataJson: "{}",
			},
			{
				Id:       anthropicEndpoint.Id,
				Protocol: protocolAnthropic,
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
	openaiEndpoint = endpointsByProtocol[protocolOpenAI]
	if len(endpoints) != 1 || openaiEndpoint == nil || openaiEndpoint.BaseUrl != "https://unit.example/openai/v2" {
		t.Fatalf("expected one updated OpenAI endpoint, got %#v", endpoints)
	}

	if _, err = svc.CreateModel(ctx, ModelSaveInput{
		ProviderId: providerID,
		EndpointId: openaiEndpoint.Id,
		ModelName:  "form-endpoint-model",
		Protocol:   protocolOpenAI,
		Enabled:    enabledYes,
	}); err != nil {
		t.Fatalf("create model referencing endpoint: %v", err)
	}
	err = svc.UpdateProvider(ctx, ProviderSaveInput{
		Id:      providerID,
		Name:    "form-endpoints-rollback",
		Enabled: enabledNo,
		Endpoints: []ProviderEndpointSaveInput{{
			Id:       openaiEndpoint.Id,
			Protocol: protocolOpenAI,
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
	snapshot := snapshotTierFor(t, ctx, svc, "image", "generate", tierCodeBasic)
	t.Cleanup(func() { restoreTier(t, ctx, snapshot) })

	endpointID, err := svc.CreateProviderEndpoint(ctx, ProviderEndpointSaveInput{
		ProviderId:   providerID,
		Protocol:     protocolOpenAICompatible,
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
		Protocol:   protocolOpenAICompatible,
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
		Protocol:     protocolOpenAICompatible,
		BaseUrl:      "https://unit.example/v2",
		SecretRef:    endpoints[0].SecretRef,
		Enabled:      enabledYes,
		MetadataJson: `{"region":"unit-2"}`,
	}); err != nil {
		t.Fatalf("update endpoint: %v", err)
	}

	modelID, err := svc.CreateModel(ctx, ModelSaveInput{
		ProviderId: providerID,
		EndpointId: endpointID,
		ModelName:  "unit-image-model",
		Protocol:   protocolOpenAICompatible,
		Enabled:    enabledYes,
	})
	if err != nil {
		t.Fatalf("create image model: %v", err)
	}
	capabilities, err := svc.ListModelCapabilities(ctx, modelID)
	if err != nil {
		t.Fatalf("list model capabilities: %v", err)
	}
	if len(capabilities) != 0 {
		t.Fatalf("model creation must not declare image capability, got %#v", capabilities)
	}
	imageModels, err := svc.ListModels(ctx, ModelListInput{
		ProviderId: providerID,
		PageNum:    1,
		PageSize:   10,
		Enabled:    intPtr(enabledYes),
	})
	if err != nil {
		t.Fatalf("list image models: %v", err)
	}
	if imageModels.Total != 1 ||
		imageModels.List[0].ModelName != "unit-image-model" {
		t.Fatalf("expected image model identity projection, got %#v", imageModels)
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
		Enabled:           enabledYes,
		SupportsStreaming: enabledNo,
	}}); err != nil {
		t.Fatalf("upsert model state: %v", err)
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
		Code:             tierCodeBasic,
		ProviderId:       providerID,
		ModelId:          modelID,
		Enabled:          enabledYes,
	}); err != nil {
		t.Fatalf("bind image tier: %v", err)
	}
	if _, err = svc.resolveTierBinding(ctx, "image", "generate", tierCodeBasic); err != nil {
		t.Fatalf("resolve image tier binding: %v", err)
	}
	if err = svc.DeleteProviderEndpoint(ctx, providerID, endpointID); !bizerr.Is(err, CodeProviderEndpointInUse) {
		t.Fatalf("expected endpoint in-use protection, got %v", err)
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
		Protocol:         protocolOpenAICompatible,
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

func (testBizCtx) Current(context.Context) bizctxcap.CurrentContext {
	return bizctxcap.CurrentContext{UserID: 1, TenantID: 0, PlatformBypass: true}
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
	items map[string]*cachecap.CacheItem
}

func newMemoryCacheService() *memoryCacheService {
	return &memoryCacheService{items: make(map[string]*cachecap.CacheItem)}
}

func (s *memoryCacheService) Get(
	_ context.Context,
	namespace string,
	key string,
) (*cachecap.CacheItem, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[memoryCacheKey(namespace, key)]
	if !ok {
		return nil, false, nil
	}
	copied := *item
	return &copied, true, nil
}

func (s *memoryCacheService) GetMany(
	_ context.Context,
	in cachecap.GetManyInput,
) (*cachecap.GetManyOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	output := &cachecap.GetManyOutput{
		Items:       make(map[string]*cachecap.CacheItem, len(in.Keys)),
		MissingKeys: []string{},
	}
	for _, key := range in.Keys {
		item, ok := s.items[memoryCacheKey(in.Namespace, key)]
		if !ok {
			output.MissingKeys = append(output.MissingKeys, key)
			continue
		}
		copied := *item
		output.Items[key] = &copied
	}
	return output, nil
}

func (s *memoryCacheService) Set(
	_ context.Context,
	namespace string,
	key string,
	value string,
	ttl time.Duration,
) (*cachecap.CacheItem, error) {
	if err := requireMemoryCachePositiveTTL(ttl); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	item := &cachecap.CacheItem{
		Key:       key,
		ValueKind: cachecap.CacheValueKindString,
		Value:     value,
		ExpireAt:  memoryCacheExpireAt(ttl),
	}
	s.items[memoryCacheKey(namespace, key)] = item
	copied := *item
	return &copied, nil
}

func (s *memoryCacheService) SetMany(
	_ context.Context,
	in cachecap.SetManyInput,
) (*cachecap.SetManyOutput, error) {
	for _, input := range in.Items {
		if err := requireMemoryCachePositiveTTL(input.TTL); err != nil {
			return nil, err
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	output := &cachecap.SetManyOutput{
		Items: make(map[string]*cachecap.CacheItem, len(in.Items)),
	}
	for _, input := range in.Items {
		item := &cachecap.CacheItem{
			Key:       input.Key,
			ValueKind: cachecap.CacheValueKindString,
			Value:     input.Value,
			ExpireAt:  memoryCacheExpireAt(input.TTL),
		}
		s.items[memoryCacheKey(in.Namespace, input.Key)] = item
		copied := *item
		output.Items[input.Key] = &copied
	}
	return output, nil
}

func (s *memoryCacheService) Delete(_ context.Context, namespace string, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, memoryCacheKey(namespace, key))
	return nil
}

func (s *memoryCacheService) DeleteMany(_ context.Context, in cachecap.DeleteManyInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, key := range in.Keys {
		delete(s.items, memoryCacheKey(in.Namespace, key))
	}
	return nil
}

func (s *memoryCacheService) Incr(
	_ context.Context,
	namespace string,
	key string,
	delta int64,
	ttl time.Duration,
) (*cachecap.CacheItem, error) {
	if err := requireMemoryCachePositiveTTL(ttl); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	cacheKey := memoryCacheKey(namespace, key)
	item := s.items[cacheKey]
	if item == nil {
		item = &cachecap.CacheItem{Key: key, ValueKind: cachecap.CacheValueKindInt}
		s.items[cacheKey] = item
	}
	item.IntValue += delta
	item.ExpireAt = memoryCacheExpireAt(ttl)
	copied := *item
	return &copied, nil
}

func (s *memoryCacheService) Expire(
	_ context.Context,
	namespace string,
	key string,
	ttl time.Duration,
) (bool, *time.Time, error) {
	if err := requireMemoryCachePositiveTTL(ttl); err != nil {
		return false, nil, err
	}
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

// requireMemoryCachePositiveTTL mirrors the host cache contract for test doubles.
func requireMemoryCachePositiveTTL(ttl time.Duration) error {
	if ttl <= 0 {
		return gerror.New("test memory cache requires positive ttl")
	}
	return nil
}

func memoryCacheExpireAt(ttl time.Duration) *time.Time {
	if ttl <= 0 {
		return nil
	}
	expireAt := time.Now().Add(ttl)
	return &expireAt
}

// assertTierRevisionExpires verifies shared tier cache revisions use a finite positive lifetime.
func assertTierRevisionExpires(t *testing.T, ctx context.Context, cacheSvc cachecap.Service) {
	t.Helper()
	item, found, err := cacheSvc.Get(ctx, tierCacheNamespace, tierCacheRevisionKey)
	if err != nil {
		t.Fatalf("read tier revision cache: %v", err)
	}
	if !found || item == nil || item.ExpireAt == nil || !item.ExpireAt.After(time.Now()) {
		t.Fatalf("expected tier revision cache item with future expiration, got found=%v item=%#v", found, item)
	}
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
		Protocol:   protocolOpenAI,
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
		ProviderId: providerID,
		EndpointId: defaultTestEndpointID(t, ctx, svc, providerID, protocolOpenAI),
		ModelName:  modelName,
		Protocol:   protocolOpenAI,
		Enabled:    enabledYes,
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

func insertInvocationFixture(t *testing.T, ctx context.Context, requestID string, createdAt time.Time) {
	t.Helper()
	if _, err := dao.Invocation.Ctx(ctx).Data(do.Invocation{
		RequestId:            requestID,
		CapabilityType:       capabilityTypeText,
		CapabilityMethod:     capabilityMethodGenerate,
		Purpose:              "unit.clean",
		TierCode:             tierCodeBasic,
		SourcePluginId:       "unit-test",
		Status:               invocationStatusSuccess,
		AssetSummaryJson:     "{}",
		OperationSummaryJson: "{}",
		MetadataSummaryJson:  "{}",
		CreatedAt:            &createdAt,
	}).Insert(); err != nil {
		t.Fatalf("insert invocation fixture: %v", err)
	}
}

func invocationExists(t *testing.T, ctx context.Context, requestID string) bool {
	t.Helper()
	count, err := dao.Invocation.Ctx(ctx).Where(do.Invocation{RequestId: requestID}).Count()
	if err != nil {
		t.Fatalf("count invocation fixture: %v", err)
	}
	return count > 0
}

func snapshotTier(t *testing.T, ctx context.Context, svc *serviceImpl, code string) tierSnapshot {
	t.Helper()
	return snapshotTierFor(t, ctx, svc, capabilityTypeText, capabilityMethodGenerate, code)
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
