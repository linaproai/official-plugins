// Package dynamicservice implements guest-side backend services for the
// linapro-demo-dynamic sample plugin.
package dynamicservice

import (
	"context"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/pluginbridge"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// Service defines the dynamic service contract.
type Service interface {
	// BuildBackendSummaryPayload builds the backend summary response payload.
	BuildBackendSummaryPayload(input *BackendSummaryInput) *backendSummaryPayload
	// ListDemoRecordsPayload returns one paged demo-record list backed by the
	// plugin-owned SQL table.
	ListDemoRecordsPayload(input *DemoRecordListInput) (*demoRecordListPayload, error)
	// GetDemoRecordPayload returns one demo-record detail by ID.
	GetDemoRecordPayload(recordID string) (*demoRecordPayload, error)
	// CreateDemoRecordPayload creates one demo record and stores its optional attachment.
	CreateDemoRecordPayload(ctx context.Context, input *DemoRecordMutationInput) (*demoRecordPayload, error)
	// UpdateDemoRecordPayload updates one demo record and replaces or removes its optional attachment.
	UpdateDemoRecordPayload(ctx context.Context, recordID string, input *DemoRecordMutationInput) (*demoRecordPayload, error)
	// DeleteDemoRecordPayload deletes one demo record and its optional attachment.
	DeleteDemoRecordPayload(ctx context.Context, recordID string) (*demoRecordDeletePayload, error)
	// BuildDemoRecordAttachmentDownload returns one attachment download descriptor.
	BuildDemoRecordAttachmentDownload(ctx context.Context, recordID string) (*demoRecordAttachmentDownloadPayload, error)
	// BuildHostCallDemoPayload executes the host service demo and returns the
	// response payload.
	BuildHostCallDemoPayload(ctx context.Context, input *HostCallDemoInput) (*hostCallDemoPayload, error)
	// BuildManifestDemoPayload reads the explicitly authorized packaged
	// manifest resources and returns the manifest host-service demo payload.
	BuildManifestDemoPayload(ctx context.Context) (*hostCallDemoManifestPayload, error)
	// RunLifecycleDebugHook logs one lifecycle callback invocation.
	RunLifecycleDebugHook(input *LifecycleDebugInput) error
	// RegisterJobs publishes built-in Jobs declarations for host-side discovery.
	RegisterJobs(plugin pluginbridge.Declarations) error
	// BuildJobHeartbeatPayload executes the declared Jobs heartbeat task.
	BuildJobHeartbeatPayload() (*jobHeartbeatPayload, error)
}

// Interface compliance assertion for the default dynamic sample service
// implementation.
var _ Service = (*serviceImpl)(nil)

// runtimeHostService abstracts guest runtime host-call helpers used by the
// sample service.
type runtimeHostService interface {
	// Log writes one structured runtime log entry through the host.
	Log(level int, message string, fields map[string]string) error
	// StateGetInt reads one integer runtime state value.
	StateGetInt(key string) (int, bool, error)
	// StateSetInt writes one integer runtime state value.
	StateSetInt(key string, value int) error
	// Now returns the current host time string.
	Now() (string, error)
	// UUID returns one host-generated unique identifier string.
	UUID() (string, error)
	// Node returns the current host node identity string.
	Node() (string, error)
}

// networkHostService abstracts guest outbound HTTP host-call helpers used by
// the sample service.
type networkHostService interface {
	// Request executes one governed outbound network request through the host.
	Request(targetURL string, request *protocol.HostServiceNetworkRequest) (*protocol.HostServiceNetworkResponse, error)
}

// pluginConfigService abstracts guest plugin config helpers used by the sample
// service.
type pluginConfigService interface {
	// Exists reports whether one plugin-owned runtime config value exists.
	Exists(ctx context.Context, key string) (bool, error)
	// String reads one plugin-owned runtime config value as a string.
	String(ctx context.Context, key string, defaultValue string) (string, error)
	// Bool reads one plugin-owned runtime config value as a bool.
	Bool(ctx context.Context, key string, defaultValue bool) (bool, error)
}

// orgHostService abstracts guest organization capability calls used by the
// sample service.
type orgHostService interface {
	// Status returns the current organization capability activation state.
	Status(ctx context.Context) capmodel.CapabilityStatus
	// Available reports whether the organization capability has an active provider.
	Available(ctx context.Context) bool
	// ListUserDeptAssignments returns user-to-department projections for the provided users.
	ListUserDeptAssignments(ctx context.Context, userIDs []int) (map[int]*orgcap.UserDeptAssignment, error)
	// GetUserDeptIDs returns one user's department identifiers.
	GetUserDeptIDs(ctx context.Context, userID int) ([]int, error)
	// GetUserPostIDs returns one user's post identifiers.
	GetUserPostIDs(ctx context.Context, userID int) ([]int, error)
}

// tenantHostService abstracts guest tenant capability calls used by the sample
// service.
type tenantHostService interface {
	// Status returns the current tenant capability activation state.
	Status(ctx context.Context) capmodel.CapabilityStatus
	// Available reports whether the tenant capability has an active provider.
	Available(ctx context.Context) bool
	// Current returns the current request tenant.
	Current(ctx context.Context) tenantcap.TenantID
	// PlatformBypass reports whether the current request may bypass tenant filtering.
	PlatformBypass(ctx context.Context) bool
	// EnsureTenantVisible validates that the current user can access tenantID.
	EnsureTenantVisible(ctx context.Context, tenantID tenantcap.TenantID) error
	// ListUserTenants returns active tenants visible to one user.
	ListUserTenants(ctx context.Context, userID int) ([]tenantcap.TenantInfo, error)
}

// serviceImpl implements Service.
type serviceImpl struct {
	runtimeSvc      runtimeHostService
	storageSvc      storagecap.Service
	networkSvc      networkHostService
	pluginConfigSvc pluginConfigService
	manifestSvc     manifestcap.Service
	hostConfigSvc   hostconfigcap.Service
	orgSvc          orgHostService
	tenantSvc       tenantHostService
	recordStoreSvc  recordStoreService
}

// New creates and returns a new dynamic plugin backend service.
func New() Service {
	return &serviceImpl{
		runtimeSvc:      newRuntimeHostService(),
		storageSvc:      newStorageHostService(),
		networkSvc:      newNetworkHostService(),
		pluginConfigSvc: newPluginConfigService(),
		manifestSvc:     newManifestHostService(),
		hostConfigSvc:   newHostConfigHostService(),
		orgSvc:          newOrgHostService(),
		tenantSvc:       newTenantHostService(),
		recordStoreSvc:  newRecordStoreService(),
	}
}
