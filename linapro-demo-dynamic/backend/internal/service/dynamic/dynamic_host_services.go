// This file binds the sample plugin service to the guest-side capability
// host-service clients. The framework guest SDK provides real clients for
// wasip1 builds and unsupported stubs for ordinary Go test builds.

package dynamicservice

import (
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	"lina-core/pkg/plugin/capability/recordstore"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/pluginbridge"
)

var guestServices = pluginbridge.Default()

// recordStoreService abstracts the governed record store facade used by the sample service.
type recordStoreService interface {
	// Table starts one single-table governed record store query builder.
	Table(table string) *recordstore.Query
	// Transaction executes one governed structured mutation transaction.
	Transaction(fn func(tx *recordstore.Tx) error) error
}

// newRuntimeHostService returns the guest-side runtime host client.
func newRuntimeHostService() runtimeHostService {
	return guestServices.Runtime()
}

// newStorageHostService returns the guest-side storage domain client.
func newStorageHostService() storagecap.Service {
	return guestServices.Storage()
}

// newNetworkHostService returns the guest-side outbound network host client.
func newNetworkHostService() networkHostService {
	return guestServices.Network()
}

// newRecordStoreService returns the guest-side governed record store facade.
func newRecordStoreService() recordStoreService {
	return guestServices.RecordStore()
}

// newPluginConfigService returns the guest-side plugin config client.
func newPluginConfigService() pluginConfigService {
	return guestServices.Plugins().Config()
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

// newOrgHostService returns the guest-side organization capability client.
func newOrgHostService() orgHostService {
	return guestServices.Org()
}

// newTenantHostService returns the guest-side tenant capability client.
func newTenantHostService() tenantHostService {
	return guestServices.Tenant()
}
