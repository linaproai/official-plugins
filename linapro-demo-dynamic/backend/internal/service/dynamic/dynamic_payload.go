// This file defines stable JSON payload models used by the dynamic sample
// plugin backend responses and demo helper records.

package dynamicservice

import (
	"bytes"
	"encoding/json"

	"github.com/gogf/gf/v2/errors/gerror"
)

// backendSummaryPayload defines the summary response returned by the dynamic
// sample backend endpoint.
type backendSummaryPayload struct {
	Message       string  `json:"message"`
	PluginID      string  `json:"pluginId"`
	PublicPath    string  `json:"publicPath"`
	Access        string  `json:"access"`
	Permission    string  `json:"permission"`
	Authenticated bool    `json:"authenticated"`
	Username      *string `json:"username,omitempty"`
	IsSuperAdmin  *bool   `json:"isSuperAdmin,omitempty"`
}

// demoRecordListPayload defines the paged list response returned by sample
// record queries.
type demoRecordListPayload struct {
	List  []*demoRecordPayload `json:"list"`
	Total int                  `json:"total"`
}

// demoRecordPayload defines one demo-record JSON payload returned by the
// dynamic sample API.
type demoRecordPayload struct {
	Id             string `json:"id"`
	Title          string `json:"title"`
	Content        string `json:"content"`
	AttachmentName string `json:"attachmentName"`
	HasAttachment  bool   `json:"hasAttachment"`
	CreatedAt      *int64 `json:"createdAt"`
	UpdatedAt      *int64 `json:"updatedAt"`
}

// demoRecordDeletePayload defines the delete response returned by sample CRUD
// endpoints.
type demoRecordDeletePayload struct {
	Id      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

// demoRecordAttachmentDownloadPayload describes one attachment download
// response before controller serialization.
type demoRecordAttachmentDownloadPayload struct {
	OriginalName string
	ContentType  string
	Body         []byte
}

// demoRecordEntity is the internal sample record representation loaded from
// structured data storage.
type demoRecordEntity struct {
	Id             string `json:"id"`
	Title          string `json:"title"`
	Content        string `json:"content"`
	AttachmentName string `json:"attachmentName"`
	AttachmentPath string `json:"attachmentPath"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

// demoRecordTimestampLayouts lists database timestamp encodings observed from
// governed record store capability responses for the dynamic sample table.
const demoRecordTimestampLayoutPrefix = "2006-01-02 15:04:05"

// demoRecordCreateRecord defines the typed insert payload used for sample
// records. Keeping record store capability mutation input typed avoids Go's wasm JSON encoder
// edge cases with directly constructed map[string]any values.
type demoRecordCreateRecord struct {
	Id             string `json:"id"`
	Title          string `json:"title"`
	Content        string `json:"content"`
	AttachmentName string `json:"attachmentName"`
	AttachmentPath string `json:"attachmentPath"`
}

// demoRecordUpdateRecord defines the typed update payload used for sample
// records.
type demoRecordUpdateRecord struct {
	Title          string `json:"title"`
	Content        string `json:"content"`
	AttachmentName string `json:"attachmentName"`
	AttachmentPath string `json:"attachmentPath"`
}

// hostCallDemoPayload defines the top-level host-service demo response.
type hostCallDemoPayload struct {
	VisitCount int                         `json:"visitCount"`
	PluginID   string                      `json:"pluginId"`
	Runtime    hostCallDemoRuntimePayload  `json:"runtime"`
	Storage    hostCallDemoStoragePayload  `json:"storage"`
	Network    hostCallDemoNetworkPayload  `json:"network"`
	Data       hostCallDemoDataPayload     `json:"data"`
	Config     hostCallDemoConfigPayload   `json:"config"`
	Manifest   hostCallDemoManifestPayload `json:"manifest"`
	BizCtx     hostCallDemoBizCtxPayload   `json:"bizctx"`
	Cache      hostCallDemoCachePayload    `json:"cache"`
	Lock       hostCallDemoLockPayload     `json:"lock"`
	Org        hostCallDemoOrgPayload      `json:"org"`
	Tenant     hostCallDemoTenantPayload   `json:"tenant"`
	AI         hostCallDemoAIPayload       `json:"ai"`
	Message    string                      `json:"message"`
}

// hostCallDemoRuntimePayload summarizes runtime host-call results.
type hostCallDemoRuntimePayload struct {
	Now  *int64 `json:"now"`
	UUID string `json:"uuid"`
	Node string `json:"node"`
}

// hostCallDemoStoragePayload summarizes storage host-call results.
type hostCallDemoStoragePayload struct {
	PathPrefix            string `json:"pathPrefix"`
	ObjectPath            string `json:"objectPath"`
	Stored                bool   `json:"stored"`
	ListedCount           int    `json:"listedCount"`
	CursorListedCount     int    `json:"cursorListedCount"`
	BatchStatCount        int    `json:"batchStatCount"`
	BatchStatMissingCount int    `json:"batchStatMissingCount"`
	BatchDeleted          bool   `json:"batchDeleted"`
	Deleted               bool   `json:"deleted"`
}

// hostCallDemoStorageRecord is the JSON object persisted into governed storage
// during the demo.
type hostCallDemoStorageRecord struct {
	PluginID string `json:"pluginId"`
	DemoKey  string `json:"demoKey"`
}

// hostCallDemoDataPayload summarizes structured-data host-call results.
type hostCallDemoDataPayload struct {
	Table      string `json:"table"`
	RecordKey  string `json:"recordKey"`
	ListTotal  int    `json:"listTotal"`
	CountTotal int    `json:"countTotal"`
	Updated    bool   `json:"updated"`
	Deleted    bool   `json:"deleted"`
}

// hostCallDemoNetworkPayload summarizes outbound network host-call results.
type hostCallDemoNetworkPayload struct {
	URL         string `json:"url"`
	Skipped     bool   `json:"skipped"`
	StatusCode  int    `json:"statusCode"`
	ContentType string `json:"contentType"`
	BodyPreview string `json:"bodyPreview"`
	Error       string `json:"error"`
}

// hostCallDemoConfigPayload summarizes plugin config and public host config reads.
type hostCallDemoConfigPayload struct {
	Plugin     hostCallDemoPluginConfigPayload `json:"plugin"`
	HostConfig hostCallDemoHostConfigPayload   `json:"hostConfig"`
}

// hostCallDemoPluginConfigPayload summarizes plugin-owned config reads.
type hostCallDemoPluginConfigPayload struct {
	Greeting            string `json:"greeting"`
	GreetingFound       bool   `json:"greetingFound"`
	FeatureEnabled      bool   `json:"featureEnabled"`
	FeatureEnabledFound bool   `json:"featureEnabledFound"`
}

// hostCallDemoHostConfigPayload summarizes whitelisted public host config reads.
type hostCallDemoHostConfigPayload struct {
	WorkspaceBasePath      string `json:"workspaceBasePath"`
	WorkspaceBasePathFound bool   `json:"workspaceBasePathFound"`
	I18nDefault            string `json:"i18nDefault"`
	I18nDefaultFound       bool   `json:"i18nDefaultFound"`
	I18nEnabled            bool   `json:"i18nEnabled"`
	I18nEnabledFound       bool   `json:"i18nEnabledFound"`
}

// hostCallDemoManifestPayload summarizes manifest.get reads against
// explicitly authorized packaged manifest resources.
type hostCallDemoManifestPayload struct {
	ProfilePath       string `json:"profilePath"`
	ProfileFound      bool   `json:"profileFound"`
	ProfileName       string `json:"profileName"`
	ProfileTier       string `json:"profileTier"`
	ProfileOwner      string `json:"profileOwner"`
	ConfigPath        string `json:"configPath"`
	ConfigFound       bool   `json:"configFound"`
	ConfigBodyPreview string `json:"configBodyPreview"`
	BatchReadCount    int    `json:"batchReadCount"`
	MissingPathCount  int    `json:"missingPathCount"`
	ListedCount       int    `json:"listedCount"`
}

// hostCallDemoManifestProfile defines the profile manifest shape read from
// manifest/config/profile.yaml.
type hostCallDemoManifestProfile struct {
	Name  string `json:"name" yaml:"name"`
	Tier  string `json:"tier" yaml:"tier"`
	Owner string `json:"owner" yaml:"owner"`
}

// hostCallDemoBizCtxPayload summarizes current request business context reads.
type hostCallDemoBizCtxPayload struct {
	UserID          int    `json:"userId"`
	Username        string `json:"username"`
	TenantID        int    `json:"tenantId"`
	PermissionCount int    `json:"permissionCount"`
	IsSuperAdmin    bool   `json:"isSuperAdmin"`
	PlatformBypass  bool   `json:"platformBypass"`
	ActingAsTenant  bool   `json:"actingAsTenant"`
}

// hostCallDemoCachePayload summarizes plugin-scoped cache capability calls.
type hostCallDemoCachePayload struct {
	Namespace        string `json:"namespace"`
	ValueKind        int    `json:"valueKind"`
	SingleFound      bool   `json:"singleFound"`
	BatchSetCount    int    `json:"batchSetCount"`
	BatchReadCount   int    `json:"batchReadCount"`
	MissingCount     int    `json:"missingCount"`
	IncrementedValue int64  `json:"incrementedValue"`
	ExpireUpdated    bool   `json:"expireUpdated"`
	Deleted          bool   `json:"deleted"`
}

// hostCallDemoLockPayload summarizes plugin-scoped lock capability calls.
type hostCallDemoLockPayload struct {
	Name         string `json:"name"`
	Acquired     bool   `json:"acquired"`
	Renewed      bool   `json:"renewed"`
	Released     bool   `json:"released"`
	TicketIssued bool   `json:"ticketIssued"`
}

// hostCallDemoOrgPayload summarizes organization capability host-service reads.
type hostCallDemoOrgPayload struct {
	Available            bool   `json:"available"`
	CapabilityID         string `json:"capabilityId"`
	ActiveProvider       string `json:"activeProvider"`
	Reason               string `json:"reason"`
	AssignmentCount      int    `json:"assignmentCount"`
	CurrentUserDeptCount int    `json:"currentUserDeptCount"`
	CurrentUserPostCount int    `json:"currentUserPostCount"`
}

// hostCallDemoTenantPayload summarizes tenant capability host-service reads.
type hostCallDemoTenantPayload struct {
	Available       bool   `json:"available"`
	CapabilityID    string `json:"capabilityId"`
	ActiveProvider  string `json:"activeProvider"`
	Reason          string `json:"reason"`
	CurrentTenantID int    `json:"currentTenantId"`
	PlatformBypass  bool   `json:"platformBypass"`
	UserTenantCount int    `json:"userTenantCount"`
	Visible         bool   `json:"visible"`
}

// hostCallDemoAIPayload summarizes owner-aware linapro-ai-core bridge reads.
type hostCallDemoAIPayload struct {
	Owner            string `json:"owner"`
	Service          string `json:"service"`
	Version          string `json:"version"`
	CapabilityType   string `json:"capabilityType"`
	CapabilityMethod string `json:"capabilityMethod"`
	Available        bool   `json:"available"`
	CapabilityID     string `json:"capabilityId"`
	ActiveProvider   string `json:"activeProvider"`
	Reason           string `json:"reason"`
}

// buildRecordMap marshals one typed record into a generic map used by the
// structured-data client.
func buildRecordMap(record any) (map[string]any, error) {
	content, err := json.Marshal(record)
	if err != nil {
		return nil, gerror.Wrap(err, "marshal demo record failed")
	}

	payload := make(map[string]any)
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.UseNumber()
	if err = decoder.Decode(&payload); err != nil {
		return nil, gerror.Wrap(err, "unmarshal demo record failed")
	}
	return payload, nil
}

// parseDemoRecordEntity converts one generic structured-data row into the
// internal demoRecordEntity shape.
func parseDemoRecordEntity(record map[string]any) (*demoRecordEntity, error) {
	content, err := json.Marshal(record)
	if err != nil {
		return nil, gerror.Wrap(err, "marshal dynamic demo record failed")
	}

	entity := &demoRecordEntity{}
	if err = json.Unmarshal(content, entity); err != nil {
		return nil, gerror.Wrap(err, "unmarshal dynamic demo record failed")
	}
	return entity, nil
}
