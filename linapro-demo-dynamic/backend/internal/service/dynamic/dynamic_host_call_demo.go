// This file implements the host service demo business logic for the dynamic
// sample plugin.

package dynamicservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/capability/lockcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aitypes"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/spi"
)

// Host-call demo constants define the governed keys, paths, and sample values
// used by the dynamic plugin host-service showcase.
const (
	hostCallDemoStateKey            = "host_call_demo_visit_count"
	hostCallDemoStoragePath         = "host-call-demo/"
	hostCallDemoStoragePrefix       = "host-call-demo/"
	hostCallDemoStorageContentType  = "application/json"
	hostCallDemoNetworkURL          = "https://example.com"
	hostCallDemoNetworkMethodGet    = "GET"
	hostCallDemoDataTable           = demoRecordTable
	hostCallDemoRecordTitlePrefix   = "Host call demo"
	hostCallDemoAnonymousUser       = "anonymous"
	hostCallDemoSummaryMessage      = "Host service demo executed through runtime, storage, network, data, plugins.config.get, manifest batch/list, hostConfig, bizctx, cache, lock, org, tenant, and linapro-ai-core owner AI services."
	hostCallDemoNetworkPreview      = 120
	hostCallDemoPluginGreetingKey   = "demo.greeting"
	hostCallDemoPluginFeatureKey    = "demo.featureEnabled"
	hostCallDemoManifestConfigPath  = "config/config.yaml"
	hostCallDemoManifestProfilePath = "config/profile.yaml"
	hostCallDemoWorkspaceKey        = "workspace.basePath"
	hostCallDemoI18nDefaultKey      = "i18n.default"
	hostCallDemoI18nEnabledKey      = "i18n.enabled"
	hostCallDemoCacheNamespace      = "host-call-demo-cache"
	hostCallDemoCacheTTL            = time.Minute
	hostCallDemoCacheExpireTTL      = 2 * time.Minute
	hostCallDemoLockName            = "host-call-demo-lock"
	hostCallDemoLockLease           = 5 * time.Second
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

	storageSummary, err := s.runHostCallDemoStorage(ctx, hostCallDemoPluginID(input), uuidValue)
	if err != nil {
		return nil, err
	}
	dataSummary, err := s.runHostCallDemoData(hostCallDemoPluginID(input), uuidValue)
	if err != nil {
		return nil, err
	}
	configSummary, err := s.runHostCallDemoConfig(ctx)
	if err != nil {
		return nil, err
	}
	manifestSummary, err := s.runHostCallDemoManifest(ctx)
	if err != nil {
		return nil, err
	}
	bizCtxSummary, err := s.runHostCallDemoBizCtx(ctx)
	if err != nil {
		return nil, err
	}
	cacheSummary, err := s.runHostCallDemoCache(ctx, uuidValue)
	if err != nil {
		return nil, err
	}
	lockSummary, err := s.runHostCallDemoLock(ctx)
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
	aiSummary, err := s.runHostCallDemoAI(ctx)
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
		BizCtx:   *bizCtxSummary,
		Cache:    *cacheSummary,
		Lock:     *lockSummary,
		Org:      *orgSummary,
		Tenant:   *tenantSummary,
		AI:       *aiSummary,
		Message:  hostCallDemoSummaryMessage,
	}, nil
}

// BuildManifestDemoPayload reads the explicitly authorized packaged manifest
// resources and returns the manifest host-service demo payload.
func (s *serviceImpl) BuildManifestDemoPayload(ctx context.Context) (*hostCallDemoManifestPayload, error) {
	return s.runHostCallDemoManifest(ctx)
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
	ctx context.Context,
	pluginID string,
	demoKey string,
) (payload *hostCallDemoStoragePayload, err error) {
	objectPath := fmt.Sprintf("%s%s.json", hostCallDemoStoragePrefix, demoKey)
	body, err := json.Marshal(&hostCallDemoStorageRecord{
		PluginID: pluginID,
		DemoKey:  demoKey,
	})
	if err != nil {
		return nil, gerror.Wrap(err, "marshal storage demo request body failed")
	}
	if _, err = s.storageSvc.Put(ctx, storagecap.PutInput{
		Path:        objectPath,
		Body:        bytes.NewReader(body),
		Size:        int64(len(body)),
		ContentType: hostCallDemoStorageContentType,
		Overwrite:   true,
	}); err != nil {
		return nil, err
	}
	deleted := false
	defer func() {
		if !deleted {
			if cleanupErr := s.storageSvc.Delete(ctx, storagecap.DeleteInput{Path: objectPath}); cleanupErr != nil && err == nil {
				err = cleanupErr
			}
		}
	}()

	readOutput, err := s.storageSvc.Get(ctx, storagecap.GetInput{Path: objectPath})
	if err != nil {
		return nil, err
	}
	if readOutput == nil || !readOutput.Found {
		return nil, gerror.New("storage demo object verification failed")
	}
	readBody, err := io.ReadAll(readOutput.Body)
	if err != nil {
		return nil, err
	}
	if readOutput.Body != nil {
		if closeErr := readOutput.Body.Close(); closeErr != nil {
			return nil, closeErr
		}
	}
	if string(readBody) != string(body) {
		return nil, gerror.New("storage demo object verification failed")
	}

	listOutput, err := s.storageSvc.List(ctx, storagecap.ListInput{Prefix: hostCallDemoStoragePrefix, Limit: 10})
	if err != nil {
		return nil, err
	}
	cursorOutput, err := s.storageSvc.ListCursor(ctx, storagecap.ListCursorInput{Prefix: hostCallDemoStoragePrefix, Limit: 10})
	if err != nil {
		return nil, err
	}
	batchStatOutput, err := s.storageSvc.BatchStat(ctx, storagecap.BatchStatInput{Paths: []string{objectPath}})
	if err != nil {
		return nil, err
	}
	if err = s.storageSvc.DeleteMany(ctx, storagecap.DeleteManyInput{Paths: []string{objectPath}}); err != nil {
		return nil, err
	}
	deleted = true

	statOutput, err := s.storageSvc.Stat(ctx, storagecap.StatInput{Path: objectPath})
	if err != nil {
		return nil, err
	}
	listedCount := 0
	if listOutput != nil {
		listedCount = len(listOutput.Objects)
	}
	cursorListedCount := 0
	if cursorOutput != nil {
		cursorListedCount = len(cursorOutput.Objects)
	}
	batchStatCount := 0
	batchStatMissingCount := 0
	if batchStatOutput != nil {
		batchStatCount = len(batchStatOutput.Objects)
		batchStatMissingCount = len(batchStatOutput.MissingPaths)
	}
	return &hostCallDemoStoragePayload{
		PathPrefix:            hostCallDemoStoragePath,
		ObjectPath:            objectPath,
		Stored:                true,
		ListedCount:           listedCount,
		CursorListedCount:     cursorListedCount,
		BatchStatCount:        batchStatCount,
		BatchStatMissingCount: batchStatMissingCount,
		BatchDeleted:          true,
		Deleted:               statOutput == nil || !statOutput.Found,
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
func (s *serviceImpl) runHostCallDemoConfig(ctx context.Context) (*hostCallDemoConfigPayload, error) {
	if s.pluginsSvc == nil {
		return nil, gerror.New("plugin service is unavailable")
	}
	configSvc := s.pluginsSvc.Config()
	if configSvc == nil {
		return nil, gerror.New("plugin config service is unavailable")
	}
	if s.hostConfigSvc == nil {
		return nil, gerror.New("hostConfig service is unavailable")
	}

	greetingFound, err := configSvc.Exists(ctx, hostCallDemoPluginGreetingKey)
	if err != nil {
		return nil, err
	}
	greeting, err := configSvc.String(ctx, hostCallDemoPluginGreetingKey, "")
	if err != nil {
		return nil, err
	}
	featureEnabledFound, err := configSvc.Exists(ctx, hostCallDemoPluginFeatureKey)
	if err != nil {
		return nil, err
	}
	featureEnabled, err := configSvc.Bool(ctx, hostCallDemoPluginFeatureKey, false)
	if err != nil {
		return nil, err
	}
	workspaceBasePathFound, err := s.hostConfigSvc.Exists(ctx, hostCallDemoWorkspaceKey)
	if err != nil {
		return nil, err
	}
	workspaceBasePath, err := s.hostConfigSvc.String(ctx, hostCallDemoWorkspaceKey, "")
	if err != nil {
		return nil, err
	}
	i18nDefaultFound, err := s.hostConfigSvc.Exists(ctx, hostCallDemoI18nDefaultKey)
	if err != nil {
		return nil, err
	}
	i18nDefault, err := s.hostConfigSvc.String(ctx, hostCallDemoI18nDefaultKey, "")
	if err != nil {
		return nil, err
	}
	i18nEnabledFound, err := s.hostConfigSvc.Exists(ctx, hostCallDemoI18nEnabledKey)
	if err != nil {
		return nil, err
	}
	i18nEnabled, err := s.hostConfigSvc.Bool(ctx, hostCallDemoI18nEnabledKey, false)
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
func (s *serviceImpl) runHostCallDemoManifest(ctx context.Context) (*hostCallDemoManifestPayload, error) {
	if s.manifestSvc == nil {
		return nil, gerror.New("manifest host service is unavailable")
	}

	profile := &hostCallDemoManifestProfile{}
	profileFound, err := s.manifestSvc.Exists(ctx, hostCallDemoManifestProfilePath)
	if err != nil {
		return nil, err
	}
	err = s.manifestSvc.Scan(ctx, hostCallDemoManifestProfilePath, "profile", profile)
	if err != nil {
		return nil, err
	}
	configContent, err := s.manifestSvc.Get(ctx, hostCallDemoManifestConfigPath)
	if err != nil {
		return nil, err
	}
	configFound := len(configContent) > 0
	getManyOutput, err := s.manifestSvc.GetMany(ctx, manifestcap.GetManyInput{
		Paths: []string{hostCallDemoManifestProfilePath, hostCallDemoManifestConfigPath},
	})
	if err != nil {
		return nil, err
	}
	listOutput, err := s.manifestSvc.List(ctx, manifestcap.ListInput{Prefix: "config/", Limit: 10})
	if err != nil {
		return nil, err
	}
	batchReadCount := 0
	missingPathCount := 0
	if getManyOutput != nil {
		batchReadCount = len(getManyOutput.Resources)
		missingPathCount = len(getManyOutput.MissingPaths)
	}
	listedCount := 0
	if listOutput != nil {
		listedCount = len(listOutput.Resources)
	}

	return &hostCallDemoManifestPayload{
		ProfilePath:       hostCallDemoManifestProfilePath,
		ProfileFound:      profileFound,
		ProfileName:       profile.Name,
		ProfileTier:       profile.Tier,
		ProfileOwner:      profile.Owner,
		ConfigPath:        hostCallDemoManifestConfigPath,
		ConfigFound:       configFound,
		ConfigBodyPreview: buildHostCallDemoBodyPreview(configContent),
		BatchReadCount:    batchReadCount,
		MissingPathCount:  missingPathCount,
		ListedCount:       listedCount,
	}, nil
}

// runHostCallDemoBizCtx demonstrates read-only request business context access.
func (s *serviceImpl) runHostCallDemoBizCtx(ctx context.Context) (*hostCallDemoBizCtxPayload, error) {
	if s.bizCtxSvc == nil {
		return nil, gerror.New("bizctx host service is unavailable")
	}

	current := s.bizCtxSvc.Current(ctx)
	return &hostCallDemoBizCtxPayload{
		UserID:          current.UserID,
		Username:        current.Username,
		TenantID:        current.TenantID,
		PermissionCount: len(current.Permissions),
		IsSuperAdmin:    current.IsSuperAdmin,
		PlatformBypass:  current.PlatformBypass,
		ActingAsTenant:  current.ActingAsTenant,
	}, nil
}

// runHostCallDemoCache demonstrates plugin-scoped cache read, write, batch,
// increment, expiration, and deletion operations.
func (s *serviceImpl) runHostCallDemoCache(
	ctx context.Context,
	demoKey string,
) (payload *hostCallDemoCachePayload, err error) {
	if s.cacheSvc == nil {
		return nil, gerror.New("cache host service is unavailable")
	}

	var (
		valueKey   = "value-" + demoKey
		batchKeyA  = "batch-" + demoKey + "-a"
		batchKeyB  = "batch-" + demoKey + "-b"
		deleteKey  = "delete-" + demoKey
		counterKey = "counter-" + demoKey
		allKeys    = []string{valueKey, batchKeyA, batchKeyB, deleteKey, counterKey}
	)
	defer func() {
		if cleanupErr := s.cacheSvc.DeleteMany(ctx, cachecap.DeleteManyInput{
			Namespace: hostCallDemoCacheNamespace,
			Keys:      allKeys,
		}); cleanupErr != nil && err == nil {
			err = cleanupErr
		}
	}()

	if _, err = s.cacheSvc.Set(ctx, hostCallDemoCacheNamespace, valueKey, "value-"+demoKey, hostCallDemoCacheTTL); err != nil {
		return nil, err
	}
	if _, err = s.cacheSvc.Set(ctx, hostCallDemoCacheNamespace, deleteKey, "delete-me", hostCallDemoCacheTTL); err != nil {
		return nil, err
	}
	if err = s.cacheSvc.Delete(ctx, hostCallDemoCacheNamespace, deleteKey); err != nil {
		return nil, err
	}
	setManyOutput, err := s.cacheSvc.SetMany(ctx, cachecap.SetManyInput{
		Namespace: hostCallDemoCacheNamespace,
		Items: []cachecap.SetManyItem{
			{Key: batchKeyA, Value: "batch-a", TTL: hostCallDemoCacheTTL},
			{Key: batchKeyB, Value: "batch-b", TTL: hostCallDemoCacheTTL},
		},
	})
	if err != nil {
		return nil, err
	}
	valueItem, found, err := s.cacheSvc.Get(ctx, hostCallDemoCacheNamespace, valueKey)
	if err != nil {
		return nil, err
	}
	getManyOutput, err := s.cacheSvc.GetMany(ctx, cachecap.GetManyInput{
		Namespace: hostCallDemoCacheNamespace,
		Keys:      []string{valueKey, batchKeyA, batchKeyB, "missing-" + demoKey},
	})
	if err != nil {
		return nil, err
	}
	counterItem, err := s.cacheSvc.Incr(ctx, hostCallDemoCacheNamespace, counterKey, 2, hostCallDemoCacheTTL)
	if err != nil {
		return nil, err
	}
	expireUpdated, _, err := s.cacheSvc.Expire(ctx, hostCallDemoCacheNamespace, valueKey, hostCallDemoCacheExpireTTL)
	if err != nil {
		return nil, err
	}
	if err = s.cacheSvc.DeleteMany(ctx, cachecap.DeleteManyInput{
		Namespace: hostCallDemoCacheNamespace,
		Keys:      []string{valueKey, batchKeyA, batchKeyB, counterKey},
	}); err != nil {
		return nil, err
	}
	_, foundAfterDelete, err := s.cacheSvc.Get(ctx, hostCallDemoCacheNamespace, valueKey)
	if err != nil {
		return nil, err
	}

	batchSetCount := 0
	if setManyOutput != nil {
		batchSetCount = len(setManyOutput.Items)
	}
	batchReadCount := 0
	missingCount := 0
	if getManyOutput != nil {
		batchReadCount = len(getManyOutput.Items)
		missingCount = len(getManyOutput.MissingKeys)
	}
	valueKind := 0
	if valueItem != nil {
		valueKind = valueItem.ValueKind
	}
	incrementedValue := int64(0)
	if counterItem != nil {
		incrementedValue = counterItem.IntValue
	}
	return &hostCallDemoCachePayload{
		Namespace:        hostCallDemoCacheNamespace,
		ValueKind:        valueKind,
		SingleFound:      found,
		BatchSetCount:    batchSetCount,
		BatchReadCount:   batchReadCount,
		MissingCount:     missingCount,
		IncrementedValue: incrementedValue,
		ExpireUpdated:    expireUpdated,
		Deleted:          !foundAfterDelete,
	}, nil
}

// runHostCallDemoLock demonstrates plugin-scoped lock acquire, renew, and release.
func (s *serviceImpl) runHostCallDemoLock(ctx context.Context) (payload *hostCallDemoLockPayload, err error) {
	if s.lockSvc == nil {
		return nil, gerror.New("lock host service is unavailable")
	}

	acquireOutput, err := s.lockSvc.Acquire(ctx, lockcap.AcquireInput{
		Name:  hostCallDemoLockName,
		Lease: hostCallDemoLockLease,
	})
	if err != nil {
		return nil, err
	}
	payload = &hostCallDemoLockPayload{Name: hostCallDemoLockName}
	if acquireOutput == nil || !acquireOutput.Acquired {
		return payload, nil
	}
	payload.Acquired = true
	payload.TicketIssued = strings.TrimSpace(acquireOutput.Ticket) != ""

	released := false
	defer func() {
		if !released && acquireOutput.Ticket != "" {
			if cleanupErr := s.lockSvc.Release(ctx, lockcap.ReleaseInput{
				Name:   hostCallDemoLockName,
				Ticket: acquireOutput.Ticket,
			}); cleanupErr != nil && err == nil {
				err = cleanupErr
			}
		}
	}()

	renewOutput, err := s.lockSvc.Renew(ctx, lockcap.RenewInput{
		Name:   hostCallDemoLockName,
		Ticket: acquireOutput.Ticket,
	})
	if err != nil {
		return nil, err
	}
	payload.Renewed = renewOutput != nil
	if err = s.lockSvc.Release(ctx, lockcap.ReleaseInput{
		Name:   hostCallDemoLockName,
		Ticket: acquireOutput.Ticket,
	}); err != nil {
		return nil, err
	}
	released = true
	payload.Released = true
	return payload, nil
}

// runHostCallDemoOrg demonstrates read-only organization capability calls
// through a dedicated dynamic-plugin host service.
func (s *serviceImpl) runHostCallDemoOrg(ctx context.Context, input *HostCallDemoInput) (*hostCallDemoOrgPayload, error) {
	if s.orgSvc == nil {
		return nil, gerror.New("org host service is unavailable")
	}

	status := s.orgSvc.Status(ctx)
	available := s.orgSvc.Available(ctx)

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

	assignments, err := s.orgSvc.Assignment().BatchListByUsers(ctx, []int{userID})
	if err != nil {
		return nil, err
	}
	payload.AssignmentCount = len(assignments)

	deptIDs, err := s.orgSvc.Assignment().GetUserDeptIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	payload.CurrentUserDeptCount = len(deptIDs)

	postIDs, err := s.orgSvc.Assignment().GetUserPostIDs(ctx, userID)
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

	var (
		status          = s.tenantSvc.Status(ctx)
		available       = s.tenantSvc.Available(ctx)
		currentTenantID = s.tenantSvc.Context().Current(ctx)
		platformBypass  = s.tenantSvc.Context().PlatformBypass(ctx)
	)
	var err error
	if err = s.tenantSvc.Directory().EnsureVisible(ctx, []tenantcap.TenantID{currentTenantID}); err != nil {
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

	tenants, err := s.tenantSvc.Membership().ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	payload.UserTenantCount = len(tenants)
	return payload, nil
}

// runHostCallDemoAI demonstrates owner-aware linapro-ai-core bridge access by
// reading text generation method status without executing provider generation.
func (s *serviceImpl) runHostCallDemoAI(ctx context.Context) (*hostCallDemoAIPayload, error) {
	if s.aiSvc == nil {
		return nil, gerror.New("linapro-ai-core owner AI service is unavailable")
	}
	status := s.aiSvc.Text().MethodStatus(ctx, aitypes.CapabilityMethodTextGenerate)
	return &hostCallDemoAIPayload{
		Owner:            spi.OwnerPluginID,
		Service:          spi.ServiceAI,
		Version:          spi.VersionV1,
		CapabilityType:   string(status.CapabilityType),
		CapabilityMethod: string(status.CapabilityMethod),
		Available:        status.Available,
		CapabilityID:     status.CapabilityStatus.CapabilityID,
		ActiveProvider:   status.CapabilityStatus.ActiveProvider,
		Reason:           status.Reason,
	}, nil
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
