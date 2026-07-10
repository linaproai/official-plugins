// Package dynamicservice implements guest-side backend services for the
// linapro-demo-dynamic sample plugin.
package dynamicservice

import (
	"context"

	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/jobcap"
	"lina-core/pkg/plugin/capability/lockcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/pluginbridge"
	"lina-core/pkg/plugin/pluginbridge/recordstore"
	"lina-plugin-linapro-ai-core/backend/cap/aicap"
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

// serviceImpl implements Service.
type serviceImpl struct {
	runtimeSvc     pluginbridge.RuntimeHostService
	storageSvc     storagecap.Service
	networkSvc     pluginbridge.NetworkHostService
	pluginsSvc     plugincap.Service
	manifestSvc    manifestcap.Service
	hostConfigSvc  hostconfigcap.Service
	bizCtxSvc      bizctxcap.Service
	cacheSvc       cachecap.Service
	lockSvc        lockcap.Service
	jobsSvc        jobcap.Service
	orgSvc         orgcap.Service
	tenantSvc      tenantcap.Service
	aiSvc          aicap.Service
	recordStoreSvc *recordstore.DB
}

// New creates and returns a new dynamic plugin backend service.
func New() Service {
	return &serviceImpl{
		runtimeSvc:     newRuntimeHostService(),
		storageSvc:     newStorageHostService(),
		networkSvc:     newNetworkHostService(),
		pluginsSvc:     newPluginsHostService(),
		manifestSvc:    newManifestHostService(),
		hostConfigSvc:  newHostConfigHostService(),
		bizCtxSvc:      newBizCtxHostService(),
		cacheSvc:       newCacheHostService(),
		lockSvc:        newLockHostService(),
		jobsSvc:        newJobsHostService(),
		orgSvc:         newOrgHostService(),
		tenantSvc:      newTenantHostService(),
		aiSvc:          newAIHostService(),
		recordStoreSvc: newRecordStoreService(),
	}
}
