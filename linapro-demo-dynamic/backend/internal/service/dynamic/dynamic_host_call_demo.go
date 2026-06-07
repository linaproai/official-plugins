// This file implements the host service demo business logic for the dynamic
// sample plugin.

package dynamicservice

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// Host-call demo constants define the governed keys, paths, and sample values
// used by the dynamic plugin host-service showcase.
const (
	hostCallDemoStateKey            = "host_call_demo_visit_count"
	hostCallDemoStoragePath         = "host-call-demo/"
	hostCallDemoStoragePrefix       = "host-call-demo"
	hostCallDemoStorageContentType  = "application/json"
	hostCallDemoNetworkURL          = "https://example.com"
	hostCallDemoNetworkMethodGet    = "GET"
	hostCallDemoDataTable           = demoRecordTable
	hostCallDemoRecordTitlePrefix   = "Host call demo"
	hostCallDemoAnonymousUser       = "anonymous"
	hostCallDemoSummaryMessage      = "Host service demo executed through runtime, storage, network, data, config, manifest, hostConfig, org, and tenant services."
	hostCallDemoNetworkPreview      = 120
	hostCallDemoPluginGreetingKey   = "demo.greeting"
	hostCallDemoPluginFeatureKey    = "demo.featureEnabled"
	hostCallDemoManifestConfigPath  = "config/config.yaml"
	hostCallDemoManifestProfilePath = "config/profile.yaml"
	hostCallDemoWorkspaceKey        = "workspace.basePath"
	hostCallDemoI18nDefaultKey      = "i18n.default"
	hostCallDemoI18nEnabledKey      = "i18n.enabled"
)

// BuildHostCallDemoPayload executes the host service demo and returns the
// response payload.
func (s *serviceImpl) BuildHostCallDemoPayload(ctx context.Context, input *HostCallDemoInput) (*hostCallDemoPayload, error) {
	nowValue, err := s.runtimeSvc.Now()
	if err != nil {
		return nil, err
	}
	uuidValue, err := s.runtimeSvc.UUID()
	if err != nil {
		return nil, err
	}
	nodeValue, err := s.runtimeSvc.Node()
	if err != nil {
		return nil, err
	}
	if err = s.runtimeSvc.Log(
		int(protocol.LogLevelInfo),
		"host service demo invoked",
		nil,
	); err != nil {
		return nil, err
	}

	visitCount, found, err := s.runtimeSvc.StateGetInt(hostCallDemoStateKey)
	if err != nil || !found {
		visitCount = 0
	}
	visitCount++
	if err = s.runtimeSvc.StateSetInt(hostCallDemoStateKey, visitCount); err != nil {
		return nil, err
	}

	storageSummary, err := s.runHostCallDemoStorage(hostCallDemoPluginID(input), uuidValue)
	if err != nil {
		return nil, err
	}
	dataSummary, err := s.runHostCallDemoData(hostCallDemoPluginID(input), uuidValue)
	if err != nil {
		return nil, err
	}
	configSummary, err := s.runHostCallDemoConfig()
	if err != nil {
		return nil, err
	}
	manifestSummary, err := s.runHostCallDemoManifest()
	if err != nil {
		return nil, err
	}
	orgSummary, err := s.runHostCallDemoOrg(ctx, input)
	if err != nil {
		return nil, err
	}
	tenantSummary, err := s.runHostCallDemoTenant(ctx, input)
	if err != nil {
		return nil, err
	}
	networkSummary := s.runHostCallDemoNetwork(input, uuidValue)

	return &hostCallDemoPayload{
		VisitCount: visitCount,
		PluginID:   hostCallDemoPluginID(input),
		Runtime: hostCallDemoRuntimePayload{
			Now:  parseHostCallDemoRuntimeNow(nowValue),
			UUID: uuidValue,
			Node: nodeValue,
		},
		Storage:  *storageSummary,
		Network:  *networkSummary,
		Data:     *dataSummary,
		Config:   *configSummary,
		Manifest: *manifestSummary,
		Org:      *orgSummary,
		Tenant:   *tenantSummary,
		Message:  hostCallDemoSummaryMessage,
	}, nil
}

// BuildManifestDemoPayload reads the explicitly authorized packaged manifest
// resources and returns the manifest host-service demo payload.
func (s *serviceImpl) BuildManifestDemoPayload() (*hostCallDemoManifestPayload, error) {
	return s.runHostCallDemoManifest()
}

// parseHostCallDemoRuntimeNow converts the runtime.info.now host-service value
// into the public Unix-millisecond API shape without using time parsers inside
// the guest Wasm module.
func parseHostCallDemoRuntimeNow(value string) *int64 {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	millis, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return nil
	}
	return &millis
}

// runHostCallDemoStorage exercises governed storage APIs and summarizes the
// round-trip result.
func (s *serviceImpl) runHostCallDemoStorage(
	pluginID string,
	demoKey string,
) (payload *hostCallDemoStoragePayload, err error) {
	objectPath := fmt.Sprintf("%s/%s.json", hostCallDemoStoragePrefix, demoKey)
	body, err := json.Marshal(&hostCallDemoStorageRecord{
		PluginID: pluginID,
		DemoKey:  demoKey,
	})
	if err != nil {
		return nil, gerror.Wrap(err, "marshal storage demo request body failed")
	}
	if _, err = s.storageSvc.Put(objectPath, body, hostCallDemoStorageContentType, true); err != nil {
		return nil, err
	}
	deleted := false
	defer func() {
		if !deleted {
			if cleanupErr := s.storageSvc.Delete(objectPath); cleanupErr != nil && err == nil {
				err = cleanupErr
			}
		}
	}()

	readBody, _, found, err := s.storageSvc.Get(objectPath)
	if err != nil {
		return nil, err
	}
	if !found || string(readBody) != string(body) {
		return nil, gerror.New("storage demo object verification failed")
	}

	objects, err := s.storageSvc.List(hostCallDemoStoragePrefix, 10)
	if err != nil {
		return nil, err
	}
	if err = s.storageSvc.Delete(objectPath); err != nil {
		return nil, err
	}
	deleted = true

	_, statFound, err := s.storageSvc.Stat(objectPath)
	if err != nil {
		return nil, err
	}
	return &hostCallDemoStoragePayload{
		PathPrefix:  hostCallDemoStoragePath,
		ObjectPath:  objectPath,
		Stored:      true,
		ListedCount: len(objects),
		Deleted:     !statFound,
	}, nil
}

// runHostCallDemoData exercises governed structured-data APIs and summarizes
// the create/list/update/delete flow.
func (s *serviceImpl) runHostCallDemoData(
	pluginID string,
	demoKey string,
) (payload *hostCallDemoDataPayload, err error) {
	recordID := "host-call-demo-" + demoKey
	createRecord, err := buildRecordMap(&demoRecordCreateRecord{
		Id:             recordID,
		Title:          hostCallDemoRecordTitlePrefix + " " + demoKey,
		Content:        "Temporary plugin-owned record created by " + pluginID + " host-call demo.",
		AttachmentName: "",
		AttachmentPath: "",
	})
	if err != nil {
		return nil, err
	}
	createResult, err := s.recordStoreSvc.Table(hostCallDemoDataTable).Insert(createRecord)
	if err != nil {
		return nil, err
	}
	if createResult == nil || createResult.Key == nil {
		return nil, gerror.New("data demo create did not return a record key")
	}

	recordKey := createResult.Key
	deleted := false
	defer func() {
		if !deleted {
			if _, cleanupErr := s.recordStoreSvc.Table(hostCallDemoDataTable).WhereKey(recordKey).Delete(); cleanupErr != nil && err == nil {
				err = cleanupErr
			}
		}
	}()

	listRecords, listTotal, err := s.recordStoreSvc.Table(hostCallDemoDataTable).
		Fields("id", "title", "content").
		WhereEq("id", recordID).
		WhereLike("title", hostCallDemoRecordTitlePrefix).
		OrderDesc("id").
		Page(1, 10).
		All()
	if err != nil {
		return nil, err
	}
	if listTotal < 1 || len(listRecords) == 0 {
		return nil, gerror.New("data demo list did not find the created record")
	}
	countTotal, err := s.recordStoreSvc.Table(hostCallDemoDataTable).
		WhereEq("id", recordID).
		WhereLike("title", hostCallDemoRecordTitlePrefix).
		Count()
	if err != nil {
		return nil, err
	}
	updateRecord, err := buildRecordMap(&demoRecordUpdateRecord{
		Title:          hostCallDemoRecordTitlePrefix + " updated " + demoKey,
		Content:        "Updated temporary plugin-owned record created by " + pluginID + " host-call demo.",
		AttachmentName: "",
		AttachmentPath: "",
	})
	if err != nil {
		return nil, err
	}
	if _, err = s.recordStoreSvc.Table(hostCallDemoDataTable).WhereKey(recordKey).Update(updateRecord); err != nil {
		return nil, err
	}

	if _, err = s.recordStoreSvc.Table(hostCallDemoDataTable).WhereKey(recordKey).Delete(); err != nil {
		return nil, err
	}
	deleted = true

	return &hostCallDemoDataPayload{
		Table:      hostCallDemoDataTable,
		RecordKey:  fmt.Sprint(recordKey),
		ListTotal:  int(listTotal),
		CountTotal: int(countTotal),
		Updated:    true,
		Deleted:    true,
	}, nil
}

// runHostCallDemoNetwork exercises the governed outbound HTTP host service and
// captures a bounded preview of the response.
func (s *serviceImpl) runHostCallDemoNetwork(input *HostCallDemoInput, demoKey string) *hostCallDemoNetworkPayload {
	result := &hostCallDemoNetworkPayload{
		URL:         hostCallDemoNetworkURL,
		Skipped:     false,
		StatusCode:  0,
		ContentType: "",
		BodyPreview: "",
		Error:       "",
	}
	if input != nil && input.SkipNetwork {
		result.Skipped = true
		return result
	}

	response, err := s.networkSvc.Request(hostCallDemoNetworkURL, &protocol.HostServiceNetworkRequest{
		Method: hostCallDemoNetworkMethodGet,
		Headers: map[string]string{
			"x-request-id": hostCallDemoRequestID(input) + "-" + demoKey,
		},
	})
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.StatusCode = int(response.StatusCode)
	result.ContentType = response.ContentType
	result.BodyPreview = buildHostCallDemoBodyPreview(response.Body)
	return result
}

// runHostCallDemoConfig demonstrates reading plugin-owned config and
// whitelisted public host config through dynamic-plugin host services.
func (s *serviceImpl) runHostCallDemoConfig() (*hostCallDemoConfigPayload, error) {
	if s.configSvc == nil {
		return nil, gerror.New("config host service is unavailable")
	}
	if s.hostConfigSvc == nil {
		return nil, gerror.New("host config host service is unavailable")
	}

	greeting, greetingFound, err := s.configSvc.String(hostCallDemoPluginGreetingKey)
	if err != nil {
		return nil, err
	}
	featureEnabled, featureEnabledFound, err := s.configSvc.Bool(hostCallDemoPluginFeatureKey)
	if err != nil {
		return nil, err
	}
	workspaceBasePath, workspaceBasePathFound, err := s.hostConfigSvc.String(hostCallDemoWorkspaceKey)
	if err != nil {
		return nil, err
	}
	i18nDefault, i18nDefaultFound, err := s.hostConfigSvc.String(hostCallDemoI18nDefaultKey)
	if err != nil {
		return nil, err
	}
	i18nEnabled, i18nEnabledFound, err := s.hostConfigSvc.Bool(hostCallDemoI18nEnabledKey)
	if err != nil {
		return nil, err
	}

	return &hostCallDemoConfigPayload{
		Plugin: hostCallDemoPluginConfigPayload{
			Greeting:            greeting,
			GreetingFound:       greetingFound,
			FeatureEnabled:      featureEnabled,
			FeatureEnabledFound: featureEnabledFound,
		},
		HostConfig: hostCallDemoHostConfigPayload{
			WorkspaceBasePath:      workspaceBasePath,
			WorkspaceBasePathFound: workspaceBasePathFound,
			I18nDefault:            i18nDefault,
			I18nDefaultFound:       i18nDefaultFound,
			I18nEnabled:            i18nEnabled,
			I18nEnabledFound:       i18nEnabledFound,
		},
	}, nil
}

// runHostCallDemoManifest demonstrates reading the plugin's own packaged
// manifest resources through explicitly authorized manifest.get paths.
func (s *serviceImpl) runHostCallDemoManifest() (*hostCallDemoManifestPayload, error) {
	if s.manifestSvc == nil {
		return nil, gerror.New("manifest host service is unavailable")
	}

	profile := &hostCallDemoManifestProfile{}
	profileFound, err := s.manifestSvc.Scan(hostCallDemoManifestProfilePath, "profile", profile)
	if err != nil {
		return nil, err
	}
	configText, configFound, err := s.manifestSvc.GetText(hostCallDemoManifestConfigPath)
	if err != nil {
		return nil, err
	}

	return &hostCallDemoManifestPayload{
		ProfilePath:       hostCallDemoManifestProfilePath,
		ProfileFound:      profileFound,
		ProfileName:       profile.Name,
		ProfileTier:       profile.Tier,
		ProfileOwner:      profile.Owner,
		ConfigPath:        hostCallDemoManifestConfigPath,
		ConfigFound:       configFound,
		ConfigBodyPreview: buildHostCallDemoBodyPreview([]byte(configText)),
	}, nil
}

// runHostCallDemoOrg demonstrates read-only organization capability calls
// through a dedicated dynamic-plugin host service.
func (s *serviceImpl) runHostCallDemoOrg(ctx context.Context, input *HostCallDemoInput) (*hostCallDemoOrgPayload, error) {
	if s.orgSvc == nil {
		return nil, gerror.New("org host service is unavailable")
	}

	status, err := s.orgSvc.Status(ctx)
	if err != nil {
		return nil, err
	}
	available, err := s.orgSvc.Available(ctx)
	if err != nil {
		return nil, err
	}

	payload := &hostCallDemoOrgPayload{
		Available:      available,
		CapabilityID:   status.CapabilityID,
		ActiveProvider: status.ActiveProvider,
		Reason:         status.Reason,
	}
	userID := hostCallDemoUserID(input)
	if userID <= 0 {
		return payload, nil
	}

	assignments, err := s.orgSvc.ListUserDeptAssignments(ctx, []int{userID})
	if err != nil {
		return nil, err
	}
	payload.AssignmentCount = len(assignments)

	deptIDs, err := s.orgSvc.GetUserDeptIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	payload.CurrentUserDeptCount = len(deptIDs)

	postIDs, err := s.orgSvc.GetUserPostIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	payload.CurrentUserPostCount = len(postIDs)
	return payload, nil
}

// runHostCallDemoTenant demonstrates tenant capability calls through a
// dedicated dynamic-plugin host service.
func (s *serviceImpl) runHostCallDemoTenant(ctx context.Context, input *HostCallDemoInput) (*hostCallDemoTenantPayload, error) {
	if s.tenantSvc == nil {
		return nil, gerror.New("tenant host service is unavailable")
	}

	status, err := s.tenantSvc.Status(ctx)
	if err != nil {
		return nil, err
	}
	available, err := s.tenantSvc.Available(ctx)
	if err != nil {
		return nil, err
	}
	currentTenantID, err := s.tenantSvc.Current(ctx)
	if err != nil {
		return nil, err
	}
	platformBypass, err := s.tenantSvc.PlatformBypass(ctx)
	if err != nil {
		return nil, err
	}
	if err = s.tenantSvc.EnsureTenantVisible(ctx, currentTenantID); err != nil {
		return nil, err
	}

	payload := &hostCallDemoTenantPayload{
		Available:       available,
		CapabilityID:    status.CapabilityID,
		ActiveProvider:  status.ActiveProvider,
		Reason:          status.Reason,
		CurrentTenantID: int(currentTenantID),
		PlatformBypass:  platformBypass,
		Visible:         true,
	}
	userID := hostCallDemoUserID(input)
	if userID <= 0 {
		return payload, nil
	}

	tenants, err := s.tenantSvc.ListUserTenants(ctx, userID)
	if err != nil {
		return nil, err
	}
	payload.UserTenantCount = len(tenants)
	return payload, nil
}

// hostCallDemoPluginID returns the normalized plugin identifier from the input.
func hostCallDemoPluginID(input *HostCallDemoInput) string {
	if input == nil {
		return ""
	}
	return strings.TrimSpace(input.PluginID)
}

// hostCallDemoRequestID returns the normalized request identifier from the
// input.
func hostCallDemoRequestID(input *HostCallDemoInput) string {
	if input == nil {
		return ""
	}
	return strings.TrimSpace(input.RequestID)
}

// hostCallDemoRoutePath returns the normalized route path from the input.
func hostCallDemoRoutePath(input *HostCallDemoInput) string {
	if input == nil {
		return ""
	}
	return strings.TrimSpace(input.RoutePath)
}

// hostCallDemoUserID returns the authenticated user identifier from the input.
func hostCallDemoUserID(input *HostCallDemoInput) int {
	if input == nil {
		return 0
	}
	return input.UserID
}

// buildHostCallDemoBodyPreview truncates one response body to the configured
// preview length.
func buildHostCallDemoBodyPreview(body []byte) string {
	preview := strings.TrimSpace(string(body))
	if preview == "" {
		return ""
	}
	if len(preview) <= hostCallDemoNetworkPreview {
		return preview
	}
	return preview[:hostCallDemoNetworkPreview]
}
