// This file binds the sample plugin service to the guest-side capability
// host-service clients. The framework guest SDK provides real clients for
// wasip1 builds and unsupported stubs for ordinary Go test builds.

package dynamicservice

import (
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
)

var guestServices = pluginbridge.Default()

// newRuntimeHostService returns the guest-side runtime host client.
func newRuntimeHostService() pluginbridge.RuntimeHostService {
	return guestServices.Runtime()
}

// newStorageHostService returns the guest-side storage domain client.
func newStorageHostService() storagecap.Service {
	return guestServices.Storage()
}

// newNetworkHostService returns the guest-side outbound network host client.
func newNetworkHostService() pluginbridge.NetworkHostService {
	return guestServices.Network()
}

// newRecordStoreService returns the guest-side governed record store facade.
func newRecordStoreService() *recordstore.DB {
	return guestServices.RecordStore()
}

// newPluginsHostService returns the guest-side plugin-domain capability client.
func newPluginsHostService() plugincap.Service {
	return guestServices.Plugins()
}

// newManifestHostService returns the guest-side plugin manifest resource
// capability client.
func newManifestHostService() manifestcap.Service {
	return guestServices.Manifest()
}

// newHostConfigHostService returns the guest-side public host config
// capability client.
func newHostConfigHostService() hostconfigcap.Service {
	return guestServices.HostConfig()
}

// newBizCtxHostService returns the guest-side business context capability client.
func newBizCtxHostService() bizctxcap.Service {
	return guestServices.BizCtx()
}

// newCacheHostService returns the guest-side plugin cache capability client.
func newCacheHostService() cachecap.Service {
	return guestServices.Cache()
}

// newLockHostService returns the guest-side distributed lock capability client.
func newLockHostService() lockcap.Service {
	return guestServices.Lock()
}

// newJobsHostService returns the guest-side scheduled-job capability client.
func newJobsHostService() jobcap.Service {
	return guestServices.Jobs()
}

// newOrgHostService returns the guest-side organization capability client.
func newOrgHostService() orgcap.Service {
	return guestServices.Org()
}

// newTenantHostService returns the guest-side tenant capability client.
func newTenantHostService() tenantcap.Service {
	return guestServices.Tenant()
}
