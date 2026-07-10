// This file defines the host-call demo DTOs for the dynamic plugin sample.

package v1

import "github.com/gogf/gf/v2/frame/g"

// HostCallDemoReq is the request for invoking the host call demo endpoint.
type HostCallDemoReq struct {
	g.Meta      `path:"/host-call-demo" method:"post" tags:"Dynamic Plugin Demo" summary:"Host calling capability demonstration" dc:"Demonstrate dynamic plugin calls to runtime, storage, network, data, plugin config, packaged manifest resources, public host config, business context, cache, lock, organization, tenant, and linapro-ai-core owner AI capabilities through the unified host service model. The endpoint writes runtime logs, reads and writes isolated plugin storage including batch metadata, cursor listing, and batch deletion, accesses governed upstreams, performs structured CRUD on authorized data tables, reads plugin-owned config keys, reads explicitly authorized manifest.get, manifest.get_many, and manifest.list resources, reads whitelisted public host config keys, reads current request business context, exercises plugin-scoped cache and lock primitives, reads current organization and tenant projections, and reads owner-aware AI text method status without executing provider generation. Passing skipNetwork=1 skips external network requests for offline verification." access:"login" permission:"linapro-demo-dynamic:backend:view" operLog:"other"`
	SkipNetwork bool `json:"skipNetwork" dc:"Whether to skip external network requests: true=skip false=normal access, default is false when omitted" eg:"false"`
}

// HostCallDemoRes is the response for the host call demo endpoint.
type HostCallDemoRes struct {
	VisitCount int                      `json:"visitCount" dc:"The current cumulative number of visits, implemented through the runtime.state host service to achieve persistent counting" eg:"1"`
	PluginID   string                   `json:"pluginId" dc:"The unique identifier of the current plugin" eg:"linapro-demo-dynamic"`
	Runtime    *HostCallDemoRuntimeRes  `json:"runtime" dc:"Summary of basic information returned by the runtime hosting service" eg:"{\"now\":1776132000000,\"uuid\":\"0d63c6a3-ec9d-4e39-a14f-d9b165a21ef9\",\"node\":\"node-1\"}"`
	Storage    *HostCallDemoStorageRes  `json:"storage" dc:"Storage host service execution summary including list, cursor list, batch stat, and batch delete smoke checks" eg:"{\"pathPrefix\":\"host-call-demo/\",\"objectPath\":\"host-call-demo/demo.json\",\"stored\":true,\"listedCount\":1,\"cursorListedCount\":1,\"batchStatCount\":1,\"batchStatMissingCount\":0,\"batchDeleted\":true,\"deleted\":true}"`
	Network    *HostCallDemoNetworkRes  `json:"network" dc:"network hosting service executive summary" eg:"{\"url\":\"https://example.com\",\"skipped\":false,\"statusCode\":200,\"contentType\":\"text/html\"}"`
	Data       *HostCallDemoDataRes     `json:"data" dc:"data host service executive summary" eg:"{\"table\":\"plugin_linapro_demo_dynamic_record\",\"recordKey\":\"host-call-demo-0d63c6a3-ec9d-4e39-a14f-d9b165a21ef9\",\"listTotal\":1,\"countTotal\":1,\"updated\":true,\"deleted\":true}"`
	Config     *HostCallDemoConfigRes   `json:"config" dc:"Plugin config and whitelisted public host config read summary" eg:"{\"plugin\":{\"greeting\":\"Hello from dynamic plugin\",\"greetingFound\":true,\"featureEnabled\":true,\"featureEnabledFound\":true},\"hostConfig\":{\"workspaceBasePath\":\"/opt/linapro\",\"workspaceBasePathFound\":true,\"i18nDefault\":\"zh-CN\",\"i18nDefaultFound\":true,\"i18nEnabled\":true,\"i18nEnabledFound\":true}}"`
	Manifest   *HostCallDemoManifestRes `json:"manifest" dc:"Packaged manifest resource read summary for explicitly authorized manifest.get, manifest.get_many, and manifest.list paths" eg:"{\"profilePath\":\"config/profile.yaml\",\"profileFound\":true,\"profileName\":\"demo-dynamic-profile\",\"profileTier\":\"sample\",\"profileOwner\":\"linapro\",\"configPath\":\"config/config.yaml\",\"configFound\":true,\"configBodyPreview\":\"demo:\\n  greeting: Hello from dynamic plugin\",\"batchReadCount\":2,\"missingPathCount\":0,\"listedCount\":2}"`
	BizCtx     *HostCallDemoBizCtxRes   `json:"bizctx" dc:"Business context host service read summary for the current request" eg:"{\"userId\":1,\"username\":\"admin\",\"tenantId\":1,\"permissionCount\":3,\"isSuperAdmin\":false,\"platformBypass\":false,\"actingAsTenant\":true}"`
	Cache      *HostCallDemoCacheRes    `json:"cache" dc:"Plugin-scoped cache host service summary covering get, get_many, set, set_many, delete, delete_many, incr, and expire methods" eg:"{\"namespace\":\"host-call-demo-cache\",\"valueKind\":1,\"singleFound\":true,\"batchSetCount\":2,\"batchReadCount\":3,\"missingCount\":1,\"incrementedValue\":2,\"expireUpdated\":true,\"deleted\":true}"`
	Lock       *HostCallDemoLockRes     `json:"lock" dc:"Plugin-scoped lock host service summary covering acquire, renew, and release methods" eg:"{\"name\":\"host-call-demo-lock\",\"acquired\":true,\"renewed\":true,\"released\":true,\"ticketIssued\":true}"`
	Org        *HostCallDemoOrgRes      `json:"org" dc:"Organization capability host service read summary" eg:"{\"available\":true,\"capabilityId\":\"framework.org.v1\",\"activeProvider\":\"linapro-org-core\",\"assignmentCount\":1,\"currentUserDeptCount\":1,\"currentUserPostCount\":2}"`
	Tenant     *HostCallDemoTenantRes   `json:"tenant" dc:"Tenant capability host service read summary" eg:"{\"available\":true,\"capabilityId\":\"framework.tenant.v1\",\"activeProvider\":\"linapro-tenant-core\",\"currentTenantId\":1,\"platformBypass\":false,\"userTenantCount\":1,\"visible\":true}"`
	AI         *HostCallDemoAIRes       `json:"ai" dc:"linapro-ai-core owner-aware AI host service read summary" eg:"{\"owner\":\"linapro-ai-core\",\"service\":\"ai\",\"version\":\"v1\",\"capabilityType\":\"text\",\"capabilityMethod\":\"generate\",\"available\":false,\"capabilityId\":\"framework.ai.text.v1\",\"reason\":\"no_provider\"}"`
	Message    string                   `json:"message" dc:"Host call demonstration information" eg:"Host service demo executed through runtime, storage, network, data, plugins.config.get, manifest batch/list, hostConfig, bizctx, cache, lock, org, tenant, and linapro-ai-core owner AI services."`
}

// HostCallDemoRuntimeRes describes runtime service results.
type HostCallDemoRuntimeRes struct {
	Now  *int64 `json:"now" dc:"Host current time as Unix timestamp in milliseconds" eg:"1776132000000"`
	UUID string `json:"uuid" dc:"The unique identifier generated by the host, used for resource isolation in this demonstration" eg:"0d63c6a3-ec9d-4e39-a14f-d9b165a21ef9"`
	Node string `json:"node" dc:"Current host node ID" eg:"node-1"`
}

// HostCallDemoStorageRes describes storage service results.
type HostCallDemoStorageRes struct {
	PathPrefix            string `json:"pathPrefix" dc:"The authorized logical path prefix used this time" eg:"host-call-demo/"`
	ObjectPath            string `json:"objectPath" dc:"The logical object path written this time" eg:"host-call-demo/demo.json"`
	Stored                bool   `json:"stored" dc:"Whether the object was successfully written and read back" eg:"true"`
	ListedCount           int    `json:"listedCount" dc:"Number of objects listed by prefix" eg:"1"`
	CursorListedCount     int    `json:"cursorListedCount" dc:"Number of objects listed by prefix through cursor pagination" eg:"1"`
	BatchStatCount        int    `json:"batchStatCount" dc:"Number of object metadata records returned by explicit-path batch stat" eg:"1"`
	BatchStatMissingCount int    `json:"batchStatMissingCount" dc:"Number of requested object paths without metadata in explicit-path batch stat" eg:"0"`
	BatchDeleted          bool   `json:"batchDeleted" dc:"Whether explicit-path batch deletion completed successfully" eg:"true"`
	Deleted               bool   `json:"deleted" dc:"Whether the temporary object was successfully deleted" eg:"true"`
}

// HostCallDemoNetworkRes describes network service results.
type HostCallDemoNetworkRes struct {
	URL         string `json:"url" dc:"The target URL or URL pattern of this application and access" eg:"https://example.com"`
	Skipped     bool   `json:"skipped" dc:"Whether network requests are skipped via skipNetwork=1" eg:"false"`
	StatusCode  int    `json:"statusCode" dc:"Upstream HTTP status code, 0 on skip or failure" eg:"200"`
	ContentType string `json:"contentType" dc:"Upstream response content type" eg:"text/html"`
	BodyPreview string `json:"bodyPreview" dc:"Upstream response body preview, returning up to the first 120 characters" eg:"<!doctype html>"`
	Error       string `json:"error" dc:"Error summary if the network request fails; empty if successful or skipped" eg:"Get https://example.com: context deadline exceeded"`
}

// HostCallDemoDataRes describes data service results.
type HostCallDemoDataRes struct {
	Table      string `json:"table" dc:"The plugin-owned authorized data table name used this time" eg:"plugin_linapro_demo_dynamic_record"`
	RecordKey  string `json:"recordKey" dc:"The record primary key value returned by the host" eg:"host-call-demo-0d63c6a3-ec9d-4e39-a14f-d9b165a21ef9"`
	ListTotal  int    `json:"listTotal" dc:"Total number of records retrieved by paging according to filter conditions" eg:"1"`
	CountTotal int    `json:"countTotal" dc:"The total number geted by executing count query based on the same filter conditions" eg:"1"`
	Updated    bool   `json:"updated" dc:"Whether the update is successfully completed and the updated record is read back" eg:"true"`
	Deleted    bool   `json:"deleted" dc:"Whether the temporary record was successfully deleted" eg:"true"`
}

// HostCallDemoConfigRes describes plugin config and public host config reads.
type HostCallDemoConfigRes struct {
	Plugin     *HostCallDemoPluginConfigRes `json:"plugin" dc:"Plugin-owned runtime config values read through Plugins().Config()" eg:"{\"greeting\":\"Hello from dynamic plugin\",\"greetingFound\":true,\"featureEnabled\":true,\"featureEnabledFound\":true}"`
	HostConfig *HostCallDemoHostConfigRes   `json:"hostConfig" dc:"Whitelisted public host config values read through the hostConfig host service" eg:"{\"workspaceBasePath\":\"/opt/linapro\",\"workspaceBasePathFound\":true,\"i18nDefault\":\"zh-CN\",\"i18nDefaultFound\":true,\"i18nEnabled\":true,\"i18nEnabledFound\":true}"`
}

// HostCallDemoPluginConfigRes describes plugin-owned config values.
type HostCallDemoPluginConfigRes struct {
	Greeting            string `json:"greeting" dc:"The demo.greeting value read from the plugin runtime config.yaml file" eg:"Hello from dynamic plugin"`
	GreetingFound       bool   `json:"greetingFound" dc:"Whether demo.greeting exists in the plugin runtime config source" eg:"true"`
	FeatureEnabled      bool   `json:"featureEnabled" dc:"The demo.featureEnabled value read from the plugin runtime config.yaml file" eg:"true"`
	FeatureEnabledFound bool   `json:"featureEnabledFound" dc:"Whether demo.featureEnabled exists in the plugin runtime config source" eg:"true"`
}

// HostCallDemoHostConfigRes describes public host config values.
type HostCallDemoHostConfigRes struct {
	WorkspaceBasePath      string `json:"workspaceBasePath" dc:"The whitelisted host workspace.basePath config value" eg:"/opt/linapro"`
	WorkspaceBasePathFound bool   `json:"workspaceBasePathFound" dc:"Whether workspace.basePath exists in the public host config view" eg:"true"`
	I18nDefault            string `json:"i18nDefault" dc:"The whitelisted host i18n.default config value" eg:"zh-CN"`
	I18nDefaultFound       bool   `json:"i18nDefaultFound" dc:"Whether i18n.default exists in the public host config view" eg:"true"`
	I18nEnabled            bool   `json:"i18nEnabled" dc:"The whitelisted host i18n.enabled config value" eg:"true"`
	I18nEnabledFound       bool   `json:"i18nEnabledFound" dc:"Whether i18n.enabled exists in the public host config view" eg:"true"`
}

// HostCallDemoManifestRes describes manifest resource read results.
type HostCallDemoManifestRes struct {
	ProfilePath       string `json:"profilePath" dc:"The explicitly authorized manifest.get path used to read the packaged profile resource" eg:"config/profile.yaml"`
	ProfileFound      bool   `json:"profileFound" dc:"Whether config/profile.yaml was found through the manifest host service" eg:"true"`
	ProfileName       string `json:"profileName" dc:"The profile.name value scanned from config/profile.yaml through the manifest host service" eg:"demo-dynamic-profile"`
	ProfileTier       string `json:"profileTier" dc:"The profile.tier value scanned from config/profile.yaml through the manifest host service" eg:"sample"`
	ProfileOwner      string `json:"profileOwner" dc:"The profile.owner value scanned from config/profile.yaml through the manifest host service" eg:"linapro"`
	ConfigPath        string `json:"configPath" dc:"The explicitly authorized manifest.get path used to read the packaged config resource as raw text" eg:"config/config.yaml"`
	ConfigFound       bool   `json:"configFound" dc:"Whether config/config.yaml was found through the manifest host service" eg:"true"`
	ConfigBodyPreview string `json:"configBodyPreview" dc:"A bounded preview of the packaged config/config.yaml manifest resource body read through manifest.get" eg:"demo:\\n  greeting: Hello from dynamic plugin"`
	BatchReadCount    int    `json:"batchReadCount" dc:"Number of resources returned by manifest.get_many for explicit authorized paths" eg:"2"`
	MissingPathCount  int    `json:"missingPathCount" dc:"Number of explicit manifest.get_many paths that did not return a resource" eg:"0"`
	ListedCount       int    `json:"listedCount" dc:"Number of resource metadata entries returned by manifest.list under the config prefix" eg:"2"`
}

// HostCallDemoBizCtxRes describes business context service results.
type HostCallDemoBizCtxRes struct {
	UserID          int    `json:"userId" dc:"Current authenticated user identifier from the business context" eg:"1"`
	Username        string `json:"username" dc:"Current authenticated username from the business context" eg:"admin"`
	TenantID        int    `json:"tenantId" dc:"Current tenant identifier from the business context" eg:"1"`
	PermissionCount int    `json:"permissionCount" dc:"Number of effective permission keys carried by the business context" eg:"3"`
	IsSuperAdmin    bool   `json:"isSuperAdmin" dc:"Whether the current user bypasses normal permission checks" eg:"false"`
	PlatformBypass  bool   `json:"platformBypass" dc:"Whether the current request runs in platform scope and bypasses tenant filtering" eg:"false"`
	ActingAsTenant  bool   `json:"actingAsTenant" dc:"Whether the current request acts through a tenant view" eg:"true"`
}

// HostCallDemoCacheRes describes cache service results.
type HostCallDemoCacheRes struct {
	Namespace        string `json:"namespace" dc:"Plugin-local authorized cache namespace used for the smoke check" eg:"host-call-demo-cache"`
	ValueKind        int    `json:"valueKind" dc:"Value kind of the single cache item: 1=string 2=integer" eg:"1"`
	SingleFound      bool   `json:"singleFound" dc:"Whether the single cache get found the value written by cache.set" eg:"true"`
	BatchSetCount    int    `json:"batchSetCount" dc:"Number of entries written by cache.set_many" eg:"2"`
	BatchReadCount   int    `json:"batchReadCount" dc:"Number of entries returned by cache.get_many" eg:"3"`
	MissingCount     int    `json:"missingCount" dc:"Number of requested keys that were missing in cache.get_many" eg:"1"`
	IncrementedValue int64  `json:"incrementedValue" dc:"Integer value returned by cache.incr after applying the demo delta" eg:"2"`
	ExpireUpdated    bool   `json:"expireUpdated" dc:"Whether cache.expire updated the expiration policy for the demo item" eg:"true"`
	Deleted          bool   `json:"deleted" dc:"Whether cache.delete_many removed the temporary demo entries" eg:"true"`
}

// HostCallDemoLockRes describes lock service results.
type HostCallDemoLockRes struct {
	Name         string `json:"name" dc:"Plugin-local authorized lock name used for the smoke check" eg:"host-call-demo-lock"`
	Acquired     bool   `json:"acquired" dc:"Whether lock.acquire acquired the demo lock" eg:"true"`
	Renewed      bool   `json:"renewed" dc:"Whether lock.renew renewed the acquired demo lock" eg:"true"`
	Released     bool   `json:"released" dc:"Whether lock.release released the acquired demo lock" eg:"true"`
	TicketIssued bool   `json:"ticketIssued" dc:"Whether the host returned an opaque lock ticket after acquisition" eg:"true"`
}

// HostCallDemoOrgRes describes organization capability results.
type HostCallDemoOrgRes struct {
	Available            bool   `json:"available" dc:"Whether the organization capability currently has an active provider" eg:"true"`
	CapabilityID         string `json:"capabilityId" dc:"The organization capability identifier reported by the host" eg:"framework.org.v1"`
	ActiveProvider       string `json:"activeProvider" dc:"The active organization provider plugin identifier, empty when unavailable" eg:"linapro-org-core"`
	Reason               string `json:"reason" dc:"Diagnostic reason returned by the host when the organization capability is unavailable" eg:""`
	AssignmentCount      int    `json:"assignmentCount" dc:"Number of department assignment projections returned for the current user" eg:"1"`
	CurrentUserDeptCount int    `json:"currentUserDeptCount" dc:"Number of department IDs returned for the current user" eg:"1"`
	CurrentUserPostCount int    `json:"currentUserPostCount" dc:"Number of post IDs returned for the current user" eg:"2"`
}

// HostCallDemoTenantRes describes tenant capability results.
type HostCallDemoTenantRes struct {
	Available       bool   `json:"available" dc:"Whether the tenant capability currently has an active provider" eg:"true"`
	CapabilityID    string `json:"capabilityId" dc:"The tenant capability identifier reported by the host" eg:"framework.tenant.v1"`
	ActiveProvider  string `json:"activeProvider" dc:"The active tenant provider plugin identifier, empty when unavailable" eg:"linapro-tenant-core"`
	Reason          string `json:"reason" dc:"Diagnostic reason returned by the host when the tenant capability is unavailable" eg:""`
	CurrentTenantID int    `json:"currentTenantId" dc:"The current request tenant identifier returned by the tenant host service" eg:"1"`
	PlatformBypass  bool   `json:"platformBypass" dc:"Whether the current request may bypass tenant filtering" eg:"false"`
	UserTenantCount int    `json:"userTenantCount" dc:"Number of active tenants visible to the current user" eg:"1"`
	Visible         bool   `json:"visible" dc:"Whether the current tenant passed the tenant visibility check" eg:"true"`
}

// HostCallDemoAIRes describes owner-aware linapro-ai-core AI method status results.
type HostCallDemoAIRes struct {
	Owner            string `json:"owner" dc:"The owner plugin identifier that publishes the AI capability contract" eg:"linapro-ai-core"`
	Service          string `json:"service" dc:"The owner-aware host service key declared by the dynamic plugin" eg:"ai"`
	Version          string `json:"version" dc:"The owner AI capability protocol version declared by the dynamic plugin" eg:"v1"`
	CapabilityType   string `json:"capabilityType" dc:"The AI capability family checked by this demo" eg:"text"`
	CapabilityMethod string `json:"capabilityMethod" dc:"The AI capability method checked by this demo" eg:"generate"`
	Available        bool   `json:"available" dc:"Whether the owner AI text generation method currently has an active provider" eg:"false"`
	CapabilityID     string `json:"capabilityId" dc:"The framework capability identifier reported by the owner AI method status" eg:"framework.ai.text.v1"`
	ActiveProvider   string `json:"activeProvider" dc:"The active AI provider plugin identifier, empty when unavailable" eg:"linapro-ai-core"`
	Reason           string `json:"reason" dc:"Diagnostic reason returned by the owner AI method status when unavailable" eg:"no_provider"`
}
