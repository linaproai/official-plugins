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
	// CapabilityMethodGenerate identifies synchronous text generation.
	CapabilityMethodGenerate = string(aitext.CapabilityMethodGenerate)
	// ProtocolOpenAI identifies the OpenAI-compatible adapter.
	ProtocolOpenAI = "openai"
	// ProtocolAnthropic identifies the Anthropic-compatible adapter.
	ProtocolAnthropic = "anthropic"
	// ProtocolVoyage identifies the Voyage-compatible adapter.
	ProtocolVoyage = "voyage"
	// ProtocolOpenAICompatible identifies a generic OpenAI-compatible adapter.
	ProtocolOpenAICompatible = "openai-compatible"
	// ProtocolAnthropicCompatible identifies a generic Anthropic-compatible adapter.
	ProtocolAnthropicCompatible = "anthropic-compatible"
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
	// InvocationPurposeTierTest identifies invocation logs emitted by tier tests.
	InvocationPurposeTierTest = "linapro-ai-core.tier.test"
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
	tierTestTimeout        = 60 * time.Second
)

// Service aggregates Smart Center management, text-generation, and tier-cache
// operations while keeping each functional contract separately named.
type Service interface {
	ProviderService
	ProviderEndpointService
	ModelService
	ModelCapabilityService
	MethodDefaultService
	TierService
	InvocationService
	TextGenerationService
	TierCacheService
}

// ProviderService defines provider metadata management operations.
type ProviderService interface {
	// ListProviders returns a platform-only paged provider list with model counts
	// assembled in one batch query. It returns business errors for non-platform contexts.
	ListProviders(ctx context.Context, in ProviderListInput) (*ProviderListOutput, error)
	// GetProvider returns one provider projection with model counts, or a not-found
	// business error when the provider is absent or soft-deleted.
	GetProvider(ctx context.Context, id int64) (*ProviderItem, error)
	// CreateProvider creates one provider metadata row and optional OpenAI or
	// Anthropic endpoint rows in one transaction, then returns the generated ID.
	CreateProvider(ctx context.Context, in ProviderSaveInput) (int64, error)
	// UpdateProvider updates one provider metadata row and, when endpoint inputs
	// are present, saves OpenAI or Anthropic endpoint rows in the same transaction.
	UpdateProvider(ctx context.Context, in ProviderSaveInput) error
	// DeleteProvider soft-deletes one provider and its unreferenced models after
	// verifying no active tier binding references that provider.
	DeleteProvider(ctx context.Context, id int64) error
}

// ProviderEndpointService defines provider protocol endpoint management operations.
type ProviderEndpointService interface {
	// ListProviderEndpoints returns provider protocol endpoints using bounded provider-scoped queries.
	ListProviderEndpoints(ctx context.Context, in ProviderEndpointListInput) ([]*ProviderEndpointItem, error)
	// CreateProviderEndpoint creates one provider protocol endpoint without returning secret plaintext.
	CreateProviderEndpoint(ctx context.Context, in ProviderEndpointSaveInput) (int64, error)
	// UpdateProviderEndpoint updates one provider protocol endpoint and keeps the stored secret when masked input is supplied.
	UpdateProviderEndpoint(ctx context.Context, in ProviderEndpointSaveInput) error
	// DeleteProviderEndpoint soft-deletes one endpoint after verifying model references.
	DeleteProviderEndpoint(ctx context.Context, providerID int64, id int64) error
}

// ModelService defines provider-owned model identity and synchronization operations.
type ModelService interface {
	// ListModels returns one bounded provider-owned model page using database-side filters.
	ListModels(ctx context.Context, in ModelListInput) (*ModelListOutput, error)
	// ListAllModels returns one bounded platform model page with provider and endpoint
	// projections batch-assembled for model-dimension management.
	ListAllModels(ctx context.Context, in ModelGlobalListInput) (*ModelListOutput, error)
	// CreateModel creates one provider-owned AI model identity row.
	CreateModel(ctx context.Context, in ModelSaveInput) (int64, error)
	// UpdateModel updates one provider-owned AI model identity row.
	UpdateModel(ctx context.Context, in ModelSaveInput) error
	// DeleteModel soft-deletes all provider-local rows sharing the target model name
	// after verifying no active tier binding references any row in that group.
	DeleteModel(ctx context.Context, id int64) error
	// SyncModels imports public model metadata from enabled provider endpoints.
	// Endpoint failures are tolerated when at least one endpoint returns models.
	SyncModels(ctx context.Context, in ModelSyncInput) (*ModelSyncOutput, error)
}

// ModelCapabilityService defines explicit model method capability metadata operations.
type ModelCapabilityService interface {
	// ListModelCapabilities returns explicit method capability declarations for one model.
	ListModelCapabilities(ctx context.Context, modelID int64) ([]*ModelCapabilityItem, error)
	// UpsertModelCapabilities replaces explicit method capability declarations for one model.
	UpsertModelCapabilities(ctx context.Context, modelID int64, items []ModelCapabilitySaveInput) error
}

// MethodDefaultService defines default parameter metadata operations for capability methods.
type MethodDefaultService interface {
	// ListMethodDefaults returns method-specific default parameter projections.
	ListMethodDefaults(ctx context.Context) ([]*MethodDefaultParamItem, error)
	// UpdateMethodDefault updates one method-specific default parameter projection.
	UpdateMethodDefault(ctx context.Context, in MethodDefaultParamSaveInput) error
}

// TierService defines fixed AI tier management and saved or draft tier test operations.
type TierService interface {
	// ListTiers returns the fixed AI tier list for one capability method with primary binding
	// projections assembled through batch queries.
	ListTiers(ctx context.Context, capabilityType string, capabilityMethod string) ([]*TierItem, error)
	// UpdateTier updates one fixed text AI tier and invalidates the tier cache
	// after the database transaction commits.
	UpdateTier(ctx context.Context, in TierUpdateInput) error
	// TestTier executes a lightweight test against a saved or draft tier binding
	// without persisting draft binding changes.
	TestTier(ctx context.Context, in TierTestInput) (*TierTestOutput, error)
}

// InvocationService defines invocation and provider-operation observation operations.
type InvocationService interface {
	// ListInvocations returns masked AI invocation logs with database-side
	// filtering, sorting, and pagination.
	ListInvocations(ctx context.Context, in InvocationListInput) (*InvocationListOutput, error)
	// CleanInvocations hard-deletes masked AI invocation logs within an optional
	// creation time range and returns the number of deleted rows.
	CleanInvocations(ctx context.Context, in InvocationCleanInput) (int, error)
	// CleanupExpiredInvocations hard-deletes invocation logs older than the
	// global retention boundary. It bypasses platform request checks because it
	// is only used by plugin lifecycle governance cron jobs.
	CleanupExpiredInvocations(ctx context.Context, retentionDays int) (int, error)
	// ListProviderOperations returns masked provider operation projections with database-side filters.
	ListProviderOperations(ctx context.Context, in ProviderOperationListInput) (*ProviderOperationListOutput, error)
	// GetProviderOperation returns one provider operation projection by opaque reference.
	GetProviderOperation(ctx context.Context, operationRef string) (*ProviderOperationItem, error)
}

// TextGenerationService defines the framework-facing text AI provider operation.
type TextGenerationService interface {
	// GenerateText executes one framework text AI request through the resolved
	// tier binding, provider protocol adapter, and masked invocation logging.
	GenerateText(ctx context.Context, request aitext.ProviderRequest) (*aitext.GenerateResponse, error)
}

// TierCacheService defines explicit tier cache invalidation operations.
type TierCacheService interface {
	// InvalidateTierCache publishes a shared tier-cache revision and removes
	// local tier bindings after successful provider, model, tier, or binding mutations.
	InvalidateTierCache(ctx context.Context, capabilityType string, capabilityMethod string, tierCode string) error
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
	Id         int64
	Name       string
	WebsiteUrl string
	Remark     string
	Enabled    int
	Endpoints  []ProviderEndpointSaveInput
}

// ProviderItem defines one provider projection.
type ProviderItem struct {
	Id                   int64
	Name                 string
	WebsiteUrl           string
	Remark               string
	Enabled              int
	ModelCount           int
	EnabledModelCount    int
	Models               []*ProviderModelSummaryItem
	EndpointCount        int
	EnabledEndpointCount int
	Endpoints            []*ProviderEndpointItem
	CreatedAt            *time.Time
	UpdatedAt            *time.Time
}

// ProviderEndpointListInput defines provider endpoint list filters.
type ProviderEndpointListInput struct {
	ProviderId int64
	Protocol   string
	Enabled    *int
}

// ProviderEndpointSaveInput defines provider endpoint create/update fields.
type ProviderEndpointSaveInput struct {
	Id           int64
	ProviderId   int64
	Protocol     string
	BaseUrl      string
	SecretRef    string
	Enabled      int
	MetadataJson string
}

// ProviderEndpointItem defines one provider endpoint projection.
type ProviderEndpointItem struct {
	Id           int64
	ProviderId   int64
	Protocol     string
	BaseUrl      string
	SecretRef    string
	Enabled      int
	MetadataJson string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

// ProviderModelSummaryItem defines the compact provider-list model projection.
type ProviderModelSummaryItem struct {
	Id        int64
	ModelName string
	Protocol  string
	Enabled   int
}

// ModelListInput defines model list filters.
type ModelListInput struct {
	ProviderId int64
	PageNum    int
	PageSize   int
	Enabled    *int
}

// ModelListOutput defines a bounded provider model list.
type ModelListOutput struct {
	List  []*ModelItem
	Total int
}

// ModelGlobalListInput defines model-dimension list filters.
type ModelGlobalListInput struct {
	PageNum    int
	PageSize   int
	Keyword    string
	ProviderId int64
	Enabled    *int
}

// ModelSaveInput defines model identity fields.
type ModelSaveInput struct {
	Id         int64
	ProviderId int64
	EndpointId int64
	ModelName  string
	Protocol   string
	Source     string
	Enabled    int
}

// ModelItem defines one model projection.
type ModelItem struct {
	Id              int64
	ProviderId      int64
	ProviderName    string
	EndpointId      int64
	EndpointBaseUrl string
	ModelName       string
	Protocol        string
	Source          string
	Enabled         int
	CreatedAt       *time.Time
	UpdatedAt       *time.Time
}

// ModelCapabilitySaveInput defines one explicit model method capability declaration.
type ModelCapabilitySaveInput struct {
	Id                int64
	EndpointId        int64
	CapabilityType    string
	CapabilityMethod  string
	InputModalities   []string
	OutputModalities  []string
	MaxInputTokens    int
	MaxOutputTokens   int
	MaxInputAssets    int
	MaxOutputAssets   int
	MaxAssetBytes     int64
	SupportsStreaming int
	SupportsOperation int
	SupportsThinking  int
	SupportedEfforts  []string
	DefaultParamsJson string
	Enabled           int
}

// ModelCapabilityItem defines one explicit model method capability projection.
type ModelCapabilityItem struct {
	Id                int64
	ModelId           int64
	EndpointId        int64
	CapabilityType    string
	CapabilityMethod  string
	InputModalities   []string
	OutputModalities  []string
	MaxInputTokens    int
	MaxOutputTokens   int
	MaxInputAssets    int
	MaxOutputAssets   int
	MaxAssetBytes     int64
	SupportsStreaming int
	SupportsOperation int
	SupportsThinking  int
	SupportedEfforts  []string
	DefaultParamsJson string
	Enabled           int
	CreatedAt         *time.Time
	UpdatedAt         *time.Time
}

// MethodDefaultParamSaveInput defines one method default params update.
type MethodDefaultParamSaveInput struct {
	CapabilityType    string
	CapabilityMethod  string
	DefaultParamsJson string
	Enabled           int
}

// MethodDefaultParamItem defines one method default params projection.
type MethodDefaultParamItem struct {
	Id                int64
	CapabilityType    string
	CapabilityMethod  string
	DefaultParamsJson string
	Enabled           int
	CreatedAt         *time.Time
	UpdatedAt         *time.Time
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
	CapabilityMethod     string
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
	ProviderId   int64
	ProviderName string
	ModelId      int64
	ModelName    string
	Protocol     string
	Enabled      int
}

// TierUpdateInput defines tier update fields.
type TierUpdateInput struct {
	CapabilityType   string
	CapabilityMethod string
	Code             string
	ProviderId       int64
	ModelId          int64
	DefaultEffort    string
	Enabled          int
}

// TierTestInput defines a saved or draft tier test request.
type TierTestInput struct {
	CapabilityType   string
	CapabilityMethod string
	Code             string
	ProviderId       int64
	ModelId          int64
	ThinkingEffort   string
	MaxOutputTokens  int
	Messages         []aitext.Message
	timeout          time.Duration
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
	PageNum          int
	PageSize         int
	CapabilityType   string
	CapabilityMethod string
	Purpose          string
	TierCode         string
	Status           string
	ProviderId       int64
	ModelId          int64
	SourcePluginId   string
	StartedAt        int64
	EndedAt          int64
}

// InvocationCleanInput defines invocation log cleanup bounds.
type InvocationCleanInput struct {
	StartedAt int64
	EndedAt   int64
}

// InvocationListOutput defines a paged masked invocation log list.
type InvocationListOutput struct {
	List  []*InvocationItem
	Total int
}

// ProviderOperationListInput defines provider operation list filters.
type ProviderOperationListInput struct {
	PageNum          int
	PageSize         int
	CapabilityType   string
	CapabilityMethod string
	Purpose          string
	Status           string
	ProviderId       int64
	ModelId          int64
	SourcePluginId   string
	StartedAt        int64
	EndedAt          int64
}

// ProviderOperationListOutput defines a bounded provider operation list.
type ProviderOperationListOutput struct {
	List  []*ProviderOperationItem
	Total int
}

// ProviderOperationItem defines one provider operation projection.
type ProviderOperationItem struct {
	Id               int64
	OperationRef     string
	CapabilityType   string
	CapabilityMethod string
	Purpose          string
	SourcePluginId   string
	ProviderId       int64
	ModelId          int64
	ProviderName     string
	ModelName        string
	Protocol         string
	Status           string
	NextPollAfterMs  int64
	ExpiresAt        *time.Time
	AssetSummaryJson string
	ErrorCode        string
	ErrorSummary     string
	CreatedAt        *time.Time
	UpdatedAt        *time.Time
}

// InvocationItem defines one masked invocation log projection.
type InvocationItem struct {
	Id                   int64
	RequestId            string
	CapabilityType       string
	CapabilityMethod     string
	Purpose              string
	TierCode             string
	SourcePluginId       string
	TenantId             int
	UserId               int
	ProviderId           int64
	ModelId              int64
	ProviderName         string
	ModelName            string
	Protocol             string
	ThinkingEffort       string
	Status               string
	InputTokens          int
	OutputTokens         int
	LatencyMs            int
	AssetSummaryJson     string
	OperationSummaryJson string
	MetadataSummaryJson  string
	ErrorCode            string
	ErrorSummary         string
	CreatedAt            *time.Time
}

// tierCacheEntry stores one cached resolved tier binding.
type tierCacheEntry struct {
	value     *resolvedTierBinding
	expiresAt time.Time
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc        plugincontract.BizCtxService
	cacheSvc         plugincontract.CacheService
	httpClient       *http.Client
	cacheMu          sync.RWMutex
	tierCache        map[string]tierCacheEntry
	revision         int64
	providerURLMu    sync.RWMutex
	providerURLCache map[string]string
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
		bizCtxSvc:        bizCtxSvc,
		cacheSvc:         cacheSvc,
		httpClient:       httpClient,
		tierCache:        make(map[string]tierCacheEntry),
		providerURLCache: make(map[string]string),
	}
}
