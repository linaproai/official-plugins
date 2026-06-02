// Package ai implements the Smart Center provider, model, tier, invocation,
// cache, and text generation services for linapro-ai-core. It owns only
// plugin-local storage and publishes a narrow provider adapter for the host
// framework text AI capability.
package ai

import (
	"context"
	"net/http"
	"sync"
	"time"

	"lina-core/pkg/plugin/capability/ai/aitext"
	plugincontract "lina-core/pkg/plugin/capability/contract"
)

const (
	// CapabilityTypeText identifies the first supported AI capability family.
	CapabilityTypeText = string(aitext.CapabilityTypeText)
	// ProtocolOpenAI identifies the OpenAI-compatible adapter.
	ProtocolOpenAI = "openai"
	// ProtocolAnthropic identifies the Anthropic-compatible adapter.
	ProtocolAnthropic = "anthropic"
	// ModelSourceManual identifies a manually maintained model row.
	ModelSourceManual = "manual"
	// ModelSourceAPI identifies a model row imported from a provider API.
	ModelSourceAPI = "api"
	// TierCodeBasic identifies the basic text AI tier.
	TierCodeBasic = string(aitext.TierBasic)
	// TierCodeStandard identifies the standard text AI tier.
	TierCodeStandard = string(aitext.TierStandard)
	// TierCodeAdvanced identifies the advanced text AI tier.
	TierCodeAdvanced = string(aitext.TierAdvanced)
	// InvocationStatusSuccess identifies a successful AI invocation.
	InvocationStatusSuccess = "success"
	// InvocationStatusFailed identifies a failed AI invocation.
	InvocationStatusFailed = "failed"
)

const (
	defaultPageNum         = 1
	defaultPageSize        = 10
	maxPageSize            = 100
	primaryBindingPriority = 0
	enabledYes             = 1
	enabledNo              = 0
	tierCacheTTL           = 30 * time.Second
	tierCacheNamespace     = "tier-binding"
	tierCacheRevisionKey   = "revision"
)

// Service defines Smart Center management and text-generation operations.
type Service interface {
	// ListProviders returns a platform-only paged provider list with model counts
	// assembled in one batch query. It returns business errors for non-platform contexts.
	ListProviders(ctx context.Context, in ProviderListInput) (*ProviderListOutput, error)
	// GetProvider returns one provider projection with model counts, or a not-found
	// business error when the provider is absent or soft-deleted.
	GetProvider(ctx context.Context, id int64) (*ProviderItem, error)
	// CreateProvider creates one provider and returns its generated identifier.
	// Plaintext API keys are not returned by any response projection.
	CreateProvider(ctx context.Context, in ProviderSaveInput) (int64, error)
	// UpdateProvider updates one provider and keeps the existing secret reference
	// when the input secret reference is empty.
	UpdateProvider(ctx context.Context, in ProviderSaveInput) error
	// DeleteProvider soft-deletes one provider and its unreferenced models after
	// verifying no active tier binding references that provider.
	DeleteProvider(ctx context.Context, id int64) error
	// ListModels returns one bounded provider-owned model page using database-side filters.
	ListModels(ctx context.Context, in ModelListInput) (*ModelListOutput, error)
	// CreateModel creates one provider-owned AI model row.
	CreateModel(ctx context.Context, in ModelSaveInput) (int64, error)
	// UpdateModel updates one provider-owned AI model row.
	UpdateModel(ctx context.Context, in ModelSaveInput) error
	// DeleteModel soft-deletes one model after verifying no active tier binding references it.
	DeleteModel(ctx context.Context, id int64) error
	// SyncModels imports public model metadata from the selected provider protocol
	// while preserving existing manual and referenced models on failure.
	SyncModels(ctx context.Context, in ModelSyncInput) (*ModelSyncOutput, error)
	// ListTiers returns the fixed text AI tier list with primary binding
	// projections assembled through batch queries.
	ListTiers(ctx context.Context, capabilityType string) ([]*TierItem, error)
	// UpdateTier updates one fixed text AI tier and invalidates the tier cache
	// after the database transaction commits.
	UpdateTier(ctx context.Context, in TierUpdateInput) error
	// TestTier executes a lightweight test against a saved or draft tier binding
	// without persisting draft binding changes.
	TestTier(ctx context.Context, in TierTestInput) (*TierTestOutput, error)
	// ListInvocations returns masked AI invocation logs with database-side
	// filtering, sorting, and pagination.
	ListInvocations(ctx context.Context, in InvocationListInput) (*InvocationListOutput, error)
	// GenerateText executes one framework text AI request through the resolved
	// tier binding, provider protocol adapter, and masked invocation logging.
	GenerateText(ctx context.Context, request aitext.ProviderRequest) (*aitext.GenerateResponse, error)
	// InvalidateTierCache publishes a shared tier-cache revision and removes
	// local tier bindings after successful provider, model, tier, or binding mutations.
	InvalidateTierCache(ctx context.Context, capabilityType string, tierCode string) error
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc  plugincontract.BizCtxService
	cacheSvc   plugincontract.CacheService
	httpClient *http.Client
	cacheMu    sync.RWMutex
	tierCache  map[string]tierCacheEntry
	revision   int64
}

// New creates and returns a Smart Center service with explicit host dependencies.
func New(
	bizCtxSvc plugincontract.BizCtxService,
	cacheSvc plugincontract.CacheService,
	httpClient *http.Client,
) Service {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &serviceImpl{
		bizCtxSvc:  bizCtxSvc,
		cacheSvc:   cacheSvc,
		httpClient: httpClient,
		tierCache:  make(map[string]tierCacheEntry),
	}
}

// ProviderListInput defines provider list filters.
type ProviderListInput struct {
	PageNum  int
	PageSize int
	Keyword  string
	Enabled  *int
}

// ProviderListOutput defines a paged provider list.
type ProviderListOutput struct {
	List  []*ProviderItem
	Total int
}

// ProviderSaveInput defines provider create/update fields.
type ProviderSaveInput struct {
	Id               int64
	Name             string
	WebsiteUrl       string
	Remark           string
	OpenaiBaseUrl    string
	AnthropicBaseUrl string
	ApiKeySecretRef  string
	Enabled          int
}

// ProviderItem defines one provider projection.
type ProviderItem struct {
	Id                int64
	Name              string
	WebsiteUrl        string
	Remark            string
	OpenaiBaseUrl     string
	AnthropicBaseUrl  string
	ApiKeySecretRef   string
	Enabled           int
	ModelCount        int
	EnabledModelCount int
	CreatedAt         *time.Time
	UpdatedAt         *time.Time
}

// ModelListInput defines model list filters.
type ModelListInput struct {
	ProviderId     int64
	PageNum        int
	PageSize       int
	CapabilityType string
	Enabled        *int
}

// ModelListOutput defines a bounded provider model list.
type ModelListOutput struct {
	List  []*ModelItem
	Total int
}

// ModelSaveInput defines model create/update fields.
type ModelSaveInput struct {
	Id               int64
	ProviderId       int64
	CapabilityType   string
	ModelName        string
	Protocol         string
	Source           string
	SupportsThinking int
	SupportedEfforts []string
	MaxInputTokens   int
	MaxOutputTokens  int
	Enabled          int
}

// ModelItem defines one model projection.
type ModelItem struct {
	Id               int64
	ProviderId       int64
	CapabilityType   string
	ModelName        string
	Protocol         string
	Source           string
	SupportsThinking int
	SupportedEfforts []string
	MaxInputTokens   int
	MaxOutputTokens  int
	Enabled          int
	CreatedAt        *time.Time
	UpdatedAt        *time.Time
}

// ModelSyncInput defines model synchronization inputs.
type ModelSyncInput struct {
	ProviderId int64
	Protocol   string
}

// ModelSyncOutput defines model synchronization counts.
type ModelSyncOutput struct {
	Created int
	Kept    int
}

// TierItem defines one fixed AI tier projection.
type TierItem struct {
	Id                   int64
	CapabilityType       string
	Code                 string
	DisplayName          string
	Description          string
	DefaultEffort        string
	Enabled              int
	SortOrder            int
	Binding              *TierBindingItem
	LastTestStatus       string
	LastTestLatencyMs    int
	LastTestErrorSummary string
	LastTestAt           *time.Time
	UpdatedAt            *time.Time
}

// TierBindingItem defines one primary binding projection.
type TierBindingItem struct {
	ProviderId       int64
	ProviderName     string
	ModelId          int64
	ModelName        string
	Protocol         string
	SupportsThinking int
	SupportedEfforts []string
	Enabled          int
}

// TierUpdateInput defines tier update fields.
type TierUpdateInput struct {
	Code          string
	ProviderId    int64
	ModelId       int64
	DefaultEffort string
	Enabled       int
}

// TierTestInput defines a saved or draft tier test request.
type TierTestInput struct {
	Code            string
	ProviderId      int64
	ModelId         int64
	ThinkingEffort  string
	MaxOutputTokens int
	Messages        []aitext.Message
}

// TierTestOutput defines the tier test result projection.
type TierTestOutput struct {
	Status         string
	LatencyMs      int
	ProviderName   string
	ModelName      string
	Protocol       string
	ThinkingEffort string
	ErrorSummary   string
	TestedAt       *time.Time
}

// InvocationListInput defines invocation log filters.
type InvocationListInput struct {
	PageNum        int
	PageSize       int
	CapabilityType string
	Purpose        string
	TierCode       string
	Status         string
	ProviderId     int64
	ModelId        int64
	SourcePluginId string
	StartedAt      int64
	EndedAt        int64
}

// InvocationListOutput defines a paged masked invocation log list.
type InvocationListOutput struct {
	List  []*InvocationItem
	Total int
}

// InvocationItem defines one masked invocation log projection.
type InvocationItem struct {
	Id             int64
	RequestId      string
	CapabilityType string
	Purpose        string
	TierCode       string
	SourcePluginId string
	TenantId       int
	UserId         int
	ProviderId     int64
	ModelId        int64
	ProviderName   string
	ModelName      string
	Protocol       string
	ThinkingEffort string
	Status         string
	InputTokens    int
	OutputTokens   int
	LatencyMs      int
	ErrorCode      string
	ErrorSummary   string
	CreatedAt      *time.Time
}

// tierCacheEntry stores one cached resolved tier binding.
type tierCacheEntry struct {
	value     *resolvedTierBinding
	expiresAt time.Time
}
