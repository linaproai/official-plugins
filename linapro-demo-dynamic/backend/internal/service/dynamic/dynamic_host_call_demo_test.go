// This file verifies host-call demo helpers that do not require a running Wasm
// host service.

package dynamicservice

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/container/gvar"

	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/lockcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-plugin-linapro-ai-core/backend/cap/aicap"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aicommon"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aitext"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/spi"
)

// fakePluginHostService exposes plugin-domain capabilities used by host-call demo tests.
type fakePluginHostService struct {
	config plugincap.ConfigService
}

// Config returns the configured fake plugin config service.
func (s *fakePluginHostService) Config() plugincap.ConfigService {
	return s.config
}

// Registry is unused by host-call demo tests.
func (s *fakePluginHostService) Registry() plugincap.RegistryService {
	return nil
}

// State is unused by host-call demo tests.
func (s *fakePluginHostService) State() plugincap.StateService {
	return nil
}

// Lifecycle is unused by host-call demo tests.
func (s *fakePluginHostService) Lifecycle() plugincap.LifecycleService {
	return nil
}

// fakePluginConfigService returns deterministic plugin config values for unit
// tests.
type fakePluginConfigService struct {
	plugincap.ConfigService
	strings map[string]configStringResult
	bools   map[string]configBoolResult
}

// configStringResult stores one string config read result.
type configStringResult struct {
	value string
	found bool
}

// configBoolResult stores one bool config read result.
type configBoolResult struct {
	value bool
	found bool
}

// Exists reports whether one fake plugin config key exists.
func (s *fakePluginConfigService) Exists(_ context.Context, key string) (bool, error) {
	if result := s.strings[key]; result.found {
		return true, nil
	}
	if result := s.bools[key]; result.found {
		return true, nil
	}
	return false, nil
}

// String returns one configured fake string value.
func (s *fakePluginConfigService) String(_ context.Context, key string, defaultValue string) (string, error) {
	result := s.strings[key]
	if !result.found {
		return defaultValue, nil
	}
	return result.value, nil
}

// Bool returns one configured fake bool value.
func (s *fakePluginConfigService) Bool(_ context.Context, key string, defaultValue bool) (bool, error) {
	result := s.bools[key]
	if !result.found {
		return defaultValue, nil
	}
	return result.value, nil
}

// fakeHostConfigHostService returns deterministic public host config values for unit
// tests.
type fakeHostConfigHostService struct {
	strings map[string]configStringResult
	bools   map[string]configBoolResult
}

// SysConfig returns no governed sys_config backend in demo tests.
func (s *fakeHostConfigHostService) SysConfig() hostconfigcap.SysConfigService {
	return &fakeHostConfigSysConfigService{}
}

// Get returns one configured fake public host config raw value.
func (s *fakeHostConfigHostService) Get(_ context.Context, key string, defaultValue any) (*gvar.Var, error) {
	if result := s.strings[key]; result.found {
		return gvar.New(result.value), nil
	}
	if result := s.bools[key]; result.found {
		return gvar.New(result.value), nil
	}
	if defaultValue != nil {
		return gvar.New(defaultValue), nil
	}
	return nil, nil
}

// Exists reports whether one configured fake public host config key exists.
func (s *fakeHostConfigHostService) Exists(_ context.Context, key string) (bool, error) {
	if result := s.strings[key]; result.found {
		return true, nil
	}
	if result := s.bools[key]; result.found {
		return true, nil
	}
	return false, nil
}

// String returns one configured fake public host config string value.
func (s *fakeHostConfigHostService) String(_ context.Context, key string, defaultValue string) (string, error) {
	result := s.strings[key]
	if !result.found {
		return defaultValue, nil
	}
	return result.value, nil
}

// Bool returns one configured fake public host config bool value.
func (s *fakeHostConfigHostService) Bool(_ context.Context, key string, defaultValue bool) (bool, error) {
	result := s.bools[key]
	if !result.found {
		return defaultValue, nil
	}
	return result.value, nil
}

// Int returns the provided default value because tests do not configure ints.
func (s *fakeHostConfigHostService) Int(_ context.Context, _ string, defaultValue int) (int, error) {
	return defaultValue, nil
}

// Duration returns the provided default value because tests do not configure durations.
func (s *fakeHostConfigHostService) Duration(_ context.Context, _ string, defaultValue time.Duration) (time.Duration, error) {
	return defaultValue, nil
}

// fakeHostConfigSysConfigService returns deterministic governed sys_config values for unit tests.
type fakeHostConfigSysConfigService struct{}

// Get returns no governed sys_config projection in demo tests.
func (*fakeHostConfigSysConfigService) Get(context.Context, hostconfigcap.SysConfigKey) (*hostconfigcap.SysConfigInfo, error) {
	return nil, nil
}

// BatchGet returns opaque missing keys because tests do not configure sys_config.
func (*fakeHostConfigSysConfigService) BatchGet(_ context.Context, keys []hostconfigcap.SysConfigKey) (*capmodel.BatchResult[*hostconfigcap.SysConfigInfo, hostconfigcap.SysConfigKey], error) {
	return &capmodel.BatchResult[*hostconfigcap.SysConfigInfo, hostconfigcap.SysConfigKey]{
		Items:      map[hostconfigcap.SysConfigKey]*hostconfigcap.SysConfigInfo{},
		MissingIDs: append([]hostconfigcap.SysConfigKey(nil), keys...),
	}, nil
}

// List returns an empty sys_config page.
func (*fakeHostConfigSysConfigService) List(context.Context, hostconfigcap.ListSysConfigInput) (*capmodel.PageResult[*hostconfigcap.SysConfigInfo], error) {
	return &capmodel.PageResult[*hostconfigcap.SysConfigInfo]{Items: []*hostconfigcap.SysConfigInfo{}}, nil
}

// SetValue fails because demo host config tests are read-only.
func (*fakeHostConfigSysConfigService) SetValue(context.Context, hostconfigcap.SysConfigKey, string) error {
	return errors.New("unexpected sys_config write in host config demo test")
}

// Reset fails because demo host config tests are read-only.
func (*fakeHostConfigSysConfigService) Reset(context.Context, hostconfigcap.SysConfigKey) error {
	return errors.New("unexpected sys_config reset in host config demo test")
}

// EnsureVisible accepts sys_config key checks in demo tests.
func (*fakeHostConfigSysConfigService) EnsureVisible(context.Context, []hostconfigcap.SysConfigKey) error {
	return nil
}

// fakeManifestHostService returns deterministic manifest resources for unit
// tests.
type fakeManifestHostService struct {
	texts   map[string]manifestTextResult
	profile hostCallDemoManifestProfile
}

// manifestTextResult stores one manifest text read result.
type manifestTextResult struct {
	value string
	found bool
}

// Get returns one configured fake manifest resource.
func (s *fakeManifestHostService) Get(_ context.Context, path string) ([]byte, error) {
	result := s.texts[path]
	if !result.found {
		return nil, nil
	}
	return []byte(result.value), nil
}

// GetMany returns configured fake manifest resources for explicit paths.
func (s *fakeManifestHostService) GetMany(_ context.Context, input manifestcap.GetManyInput) (*manifestcap.GetManyOutput, error) {
	output := &manifestcap.GetManyOutput{Resources: []*manifestcap.ResourceContent{}}
	for _, path := range input.Paths {
		result := s.texts[path]
		if !result.found {
			output.MissingPaths = append(output.MissingPaths, path)
			continue
		}
		output.Resources = append(output.Resources, &manifestcap.ResourceContent{
			Path: path,
			Body: []byte(result.value),
		})
	}
	return output, nil
}

// List returns metadata for configured fake manifest resources.
func (s *fakeManifestHostService) List(_ context.Context, input manifestcap.ListInput) (*manifestcap.ListOutput, error) {
	output := &manifestcap.ListOutput{
		Resources: []*manifestcap.Resource{},
		Limit:     input.Limit,
	}
	for path, result := range s.texts {
		if !result.found || !strings.HasPrefix(path, input.Prefix) {
			continue
		}
		output.Resources = append(output.Resources, &manifestcap.Resource{
			Path: path,
			Size: int64(len(result.value)),
		})
	}
	return output, nil
}

// Exists reports whether one configured fake manifest resource exists.
func (s *fakeManifestHostService) Exists(_ context.Context, path string) (bool, error) {
	if path == hostCallDemoManifestProfilePath {
		return true, nil
	}
	return s.texts[path].found, nil
}

// Scan copies the configured profile into the target for the expected profile
// path and key.
func (s *fakeManifestHostService) Scan(_ context.Context, path string, key string, target any) error {
	if path != hostCallDemoManifestProfilePath || strings.TrimSpace(key) != "profile" {
		return nil
	}
	profile, ok := target.(*hostCallDemoManifestProfile)
	if !ok {
		return nil
	}
	*profile = s.profile
	return nil
}

// fakeStorageHostService stores objects in memory for storage smoke tests.
type fakeStorageHostService struct {
	objects        map[string]fakeStorageObject
	deleteManyCall int
	listCursorCall int
	batchStatCall  int
}

// fakeStorageObject stores one fake storage body and metadata.
type fakeStorageObject struct {
	body        []byte
	contentType string
}

// Put stores one fake object.
func (s *fakeStorageHostService) Put(_ context.Context, in storagecap.PutInput) (*storagecap.PutOutput, error) {
	body, err := io.ReadAll(in.Body)
	if err != nil {
		return nil, err
	}
	if s.objects == nil {
		s.objects = make(map[string]fakeStorageObject)
	}
	s.objects[in.Path] = fakeStorageObject{body: body, contentType: in.ContentType}
	return &storagecap.PutOutput{Object: s.objectForPath(in.Path)}, nil
}

// Get reads one fake object.
func (s *fakeStorageHostService) Get(_ context.Context, in storagecap.GetInput) (*storagecap.GetOutput, error) {
	item, ok := s.objects[in.Path]
	if !ok {
		return &storagecap.GetOutput{Found: false}, nil
	}
	return &storagecap.GetOutput{
		Object: s.objectForPath(in.Path),
		Body:   io.NopCloser(bytes.NewReader(append([]byte(nil), item.body...))),
		Found:  true,
	}, nil
}

// Delete removes one fake object.
func (s *fakeStorageHostService) Delete(_ context.Context, in storagecap.DeleteInput) error {
	delete(s.objects, in.Path)
	return nil
}

// DeleteMany removes fake objects by explicit path.
func (s *fakeStorageHostService) DeleteMany(_ context.Context, in storagecap.DeleteManyInput) error {
	s.deleteManyCall++
	for _, path := range in.Paths {
		delete(s.objects, path)
	}
	return nil
}

// List returns fake objects with the requested prefix.
func (s *fakeStorageHostService) List(_ context.Context, in storagecap.ListInput) (*storagecap.ListOutput, error) {
	objects := s.listObjects(in.Prefix, in.Limit)
	return &storagecap.ListOutput{Objects: objects, Limit: in.Limit}, nil
}

// ListCursor returns fake objects with the requested prefix.
func (s *fakeStorageHostService) ListCursor(_ context.Context, in storagecap.ListCursorInput) (*storagecap.ListCursorOutput, error) {
	s.listCursorCall++
	objects := s.listObjects(in.Prefix, in.Limit)
	return &storagecap.ListCursorOutput{Objects: objects, Limit: in.Limit}, nil
}

// Stat returns fake object metadata.
func (s *fakeStorageHostService) Stat(_ context.Context, in storagecap.StatInput) (*storagecap.StatOutput, error) {
	if _, ok := s.objects[in.Path]; !ok {
		return &storagecap.StatOutput{Found: false}, nil
	}
	return &storagecap.StatOutput{Object: s.objectForPath(in.Path), Found: true}, nil
}

// BatchStat returns fake object metadata by explicit paths.
func (s *fakeStorageHostService) BatchStat(_ context.Context, in storagecap.BatchStatInput) (*storagecap.BatchStatOutput, error) {
	s.batchStatCall++
	output := &storagecap.BatchStatOutput{Objects: []*storagecap.Object{}}
	for _, path := range in.Paths {
		if _, ok := s.objects[path]; !ok {
			output.MissingPaths = append(output.MissingPaths, path)
			continue
		}
		output.Objects = append(output.Objects, s.objectForPath(path))
	}
	return output, nil
}

// ProviderStatuses returns no fake provider statuses.
func (s *fakeStorageHostService) ProviderStatuses(context.Context) ([]*storagecap.ProviderStatus, error) {
	return nil, nil
}

func (s *fakeStorageHostService) listObjects(prefix string, limit int) []*storagecap.Object {
	if limit <= 0 {
		limit = storagecap.DefaultListLimit
	}
	objects := make([]*storagecap.Object, 0, len(s.objects))
	for path := range s.objects {
		if !strings.HasPrefix(path, prefix) {
			continue
		}
		objects = append(objects, s.objectForPath(path))
		if len(objects) >= limit {
			break
		}
	}
	return objects
}

func (s *fakeStorageHostService) objectForPath(path string) *storagecap.Object {
	item := s.objects[path]
	return &storagecap.Object{
		Path:        path,
		Size:        int64(len(item.body)),
		ContentType: item.contentType,
		Visibility:  storagecap.VisibilityPrivate,
	}
}

// fakeBizCtxHostService returns deterministic business-context values.
type fakeBizCtxHostService struct {
	current bizctxcap.CurrentContext
}

// Current returns the configured fake business context.
func (s *fakeBizCtxHostService) Current(context.Context) bizctxcap.CurrentContext {
	return s.current
}

// fakeCacheHostService stores cache items in memory for cache smoke tests.
type fakeCacheHostService struct {
	items map[string]*cachecap.CacheItem
}

// Get returns one fake cache item.
func (s *fakeCacheHostService) Get(_ context.Context, namespace string, key string) (*cachecap.CacheItem, bool, error) {
	item, ok := s.items[fakeCacheStorageKey(namespace, key)]
	if !ok {
		return nil, false, nil
	}
	return cloneCacheItem(item), true, nil
}

// GetMany returns fake cache items for explicit keys.
func (s *fakeCacheHostService) GetMany(_ context.Context, in cachecap.GetManyInput) (*cachecap.GetManyOutput, error) {
	output := &cachecap.GetManyOutput{Items: map[string]*cachecap.CacheItem{}}
	for _, key := range in.Keys {
		item, ok := s.items[fakeCacheStorageKey(in.Namespace, key)]
		if !ok {
			output.MissingKeys = append(output.MissingKeys, key)
			continue
		}
		output.Items[key] = cloneCacheItem(item)
	}
	return output, nil
}

// Set stores one fake string cache item.
func (s *fakeCacheHostService) Set(_ context.Context, namespace string, key string, value string, ttl time.Duration) (*cachecap.CacheItem, error) {
	if s.items == nil {
		s.items = make(map[string]*cachecap.CacheItem)
	}
	item := &cachecap.CacheItem{
		Key:       key,
		ValueKind: cachecap.CacheValueKindString,
		Value:     value,
		ExpireAt:  cacheExpireAt(ttl),
	}
	s.items[fakeCacheStorageKey(namespace, key)] = item
	return cloneCacheItem(item), nil
}

// SetMany stores fake string cache items.
func (s *fakeCacheHostService) SetMany(ctx context.Context, in cachecap.SetManyInput) (*cachecap.SetManyOutput, error) {
	output := &cachecap.SetManyOutput{Items: map[string]*cachecap.CacheItem{}}
	for _, item := range in.Items {
		stored, err := s.Set(ctx, in.Namespace, item.Key, item.Value, item.TTL)
		if err != nil {
			return nil, err
		}
		output.Items[item.Key] = stored
	}
	return output, nil
}

// Delete removes one fake cache item.
func (s *fakeCacheHostService) Delete(_ context.Context, namespace string, key string) error {
	delete(s.items, fakeCacheStorageKey(namespace, key))
	return nil
}

// DeleteMany removes fake cache items by explicit keys.
func (s *fakeCacheHostService) DeleteMany(_ context.Context, in cachecap.DeleteManyInput) error {
	for _, key := range in.Keys {
		delete(s.items, fakeCacheStorageKey(in.Namespace, key))
	}
	return nil
}

// Incr increments one fake integer cache item.
func (s *fakeCacheHostService) Incr(_ context.Context, namespace string, key string, delta int64, ttl time.Duration) (*cachecap.CacheItem, error) {
	if s.items == nil {
		s.items = make(map[string]*cachecap.CacheItem)
	}
	storageKey := fakeCacheStorageKey(namespace, key)
	item := s.items[storageKey]
	if item == nil || item.ValueKind != cachecap.CacheValueKindInt {
		item = &cachecap.CacheItem{Key: key, ValueKind: cachecap.CacheValueKindInt}
	}
	item.IntValue += delta
	item.ExpireAt = cacheExpireAt(ttl)
	s.items[storageKey] = item
	return cloneCacheItem(item), nil
}

// Expire updates one fake cache item expiration.
func (s *fakeCacheHostService) Expire(_ context.Context, namespace string, key string, ttl time.Duration) (bool, *time.Time, error) {
	item := s.items[fakeCacheStorageKey(namespace, key)]
	if item == nil {
		return false, nil, nil
	}
	item.ExpireAt = cacheExpireAt(ttl)
	return true, item.ExpireAt, nil
}

func fakeCacheStorageKey(namespace string, key string) string {
	return namespace + "\x00" + key
}

func cacheExpireAt(ttl time.Duration) *time.Time {
	expireAt := time.Now().Add(ttl)
	return &expireAt
}

func cloneCacheItem(item *cachecap.CacheItem) *cachecap.CacheItem {
	if item == nil {
		return nil
	}
	cloned := *item
	return &cloned
}

// fakeLockHostService stores one fake lock ticket per lock name.
type fakeLockHostService struct {
	tickets map[string]string
}

// Acquire issues one deterministic fake lock ticket.
func (s *fakeLockHostService) Acquire(_ context.Context, in lockcap.AcquireInput) (*lockcap.AcquireOutput, error) {
	if s.tickets == nil {
		s.tickets = make(map[string]string)
	}
	ticket := in.Name + "-ticket"
	s.tickets[in.Name] = ticket
	expireAt := time.Now().Add(in.Lease)
	return &lockcap.AcquireOutput{Acquired: true, Ticket: ticket, ExpireAt: &expireAt}, nil
}

// Renew renews one fake lock ticket.
func (s *fakeLockHostService) Renew(_ context.Context, in lockcap.RenewInput) (*lockcap.RenewOutput, error) {
	if s.tickets[in.Name] != in.Ticket {
		return nil, errors.New("unexpected fake lock ticket")
	}
	expireAt := time.Now().Add(hostCallDemoLockLease)
	return &lockcap.RenewOutput{ExpireAt: &expireAt}, nil
}

// Release releases one fake lock ticket.
func (s *fakeLockHostService) Release(_ context.Context, in lockcap.ReleaseInput) error {
	if s.tickets[in.Name] != in.Ticket {
		return errors.New("unexpected fake lock ticket")
	}
	delete(s.tickets, in.Name)
	return nil
}

// fakeOrgHostService returns deterministic organization capability values for
// unit tests.
type fakeOrgHostService struct{}

// fakeOrgDepartmentService returns deterministic department capability values.
type fakeOrgDepartmentService struct{}

// fakeOrgPostService returns deterministic post capability values.
type fakeOrgPostService struct{}

// fakeOrgAssignmentService returns deterministic assignment capability values.
type fakeOrgAssignmentService struct{}

// Status returns a deterministic organization capability status.
func (s *fakeOrgHostService) Status(_ context.Context) capmodel.CapabilityStatus {
	return capmodel.CapabilityStatus{
		CapabilityID:   orgcap.CapabilityOrgV1,
		Available:      true,
		ActiveProvider: orgcap.ProviderPluginID,
	}
}

// Available reports that the fake organization capability is active.
func (s *fakeOrgHostService) Available(_ context.Context) bool {
	return true
}

// Department returns the fake department subresource.
func (s *fakeOrgHostService) Department() orgcap.DepartmentService {
	return fakeOrgDepartmentService{}
}

// Post returns the fake post subresource.
func (s *fakeOrgHostService) Post() orgcap.PostService {
	return fakeOrgPostService{}
}

// Assignment returns the fake assignment subresource.
func (s *fakeOrgHostService) Assignment() orgcap.AssignmentService {
	return fakeOrgAssignmentService{}
}

// BatchListByUsers returns one deterministic current-user assignment.
func (s fakeOrgAssignmentService) BatchListByUsers(
	_ context.Context,
	userIDs []int,
) (map[int]*orgcap.UserDeptAssignment, error) {
	result := make(map[int]*orgcap.UserDeptAssignment, len(userIDs))
	for _, userID := range userIDs {
		result[userID] = &orgcap.UserDeptAssignment{DeptID: 11, DeptName: "Engineering"}
	}
	return result, nil
}

// BatchGetUserProfiles returns deterministic empty organization profiles.
func (s fakeOrgAssignmentService) BatchGetUserProfiles(_ context.Context, userIDs []int) (*capmodel.BatchResult[*orgcap.UserOrgProfile, int], error) {
	result := &capmodel.BatchResult[*orgcap.UserOrgProfile, int]{
		Items:      make(map[int]*orgcap.UserOrgProfile, len(userIDs)),
		MissingIDs: []int{},
	}
	for _, userID := range userIDs {
		result.Items[userID] = &orgcap.UserOrgProfile{UserID: userID}
	}
	return result, nil
}

// ListByUser returns one deterministic organization profile.
func (s fakeOrgAssignmentService) ListByUser(context.Context, int) (*orgcap.UserOrgProfile, error) {
	return &orgcap.UserOrgProfile{}, nil
}

// GetUserDeptInfo returns a deterministic current-user department projection.
func (s fakeOrgAssignmentService) GetUserDeptInfo(context.Context, int) (int, string, error) {
	return 11, "Engineering", nil
}

// GetUserDeptIDs returns deterministic current-user department IDs.
func (s fakeOrgAssignmentService) GetUserDeptIDs(_ context.Context, _ int) ([]int, error) {
	return []int{11}, nil
}

// GetUserPostIDs returns deterministic current-user post IDs.
func (s fakeOrgAssignmentService) GetUserPostIDs(_ context.Context, _ int) ([]int, error) {
	return []int{21, 22}, nil
}

// ReplaceByUser is unused by host-call demo tests.
func (s fakeOrgAssignmentService) ReplaceByUser(context.Context, int, *int, []int) error {
	return nil
}

// CleanupByUser is unused by host-call demo tests.
func (s fakeOrgAssignmentService) CleanupByUser(context.Context, int) error {
	return nil
}

// Get returns no fake department projection.
func (s fakeOrgDepartmentService) Get(context.Context, int) (*orgcap.DeptInfo, error) {
	return nil, nil
}

// BatchGet returns no fake department projections.
func (s fakeOrgDepartmentService) BatchGet(context.Context, []int) (*capmodel.BatchResult[*orgcap.DeptInfo, int], error) {
	return &capmodel.BatchResult[*orgcap.DeptInfo, int]{Items: map[int]*orgcap.DeptInfo{}}, nil
}

// List returns no fake department projections.
func (s fakeOrgDepartmentService) List(context.Context, orgcap.DeptListInput) (*capmodel.PageResult[*orgcap.DeptInfo], error) {
	return &capmodel.PageResult[*orgcap.DeptInfo]{Items: []*orgcap.DeptInfo{}}, nil
}

// ListTree returns no fake department tree.
func (s fakeOrgDepartmentService) ListTree(context.Context, orgcap.DeptTreeInput) (*orgcap.DeptTreeResult, error) {
	return &orgcap.DeptTreeResult{Items: []*orgcap.DeptTreeNode{}}, nil
}

// ListOptions returns no fake department options.
func (s fakeOrgDepartmentService) ListOptions(context.Context, orgcap.DeptOptionsInput) (*capmodel.PageResult[*orgcap.DeptInfo], error) {
	return &capmodel.PageResult[*orgcap.DeptInfo]{Items: []*orgcap.DeptInfo{}}, nil
}

// EnsureVisible accepts fake department identifiers.
func (s fakeOrgDepartmentService) EnsureVisible(context.Context, []int) error {
	return nil
}

// Create is unused by host-call demo tests.
func (s fakeOrgDepartmentService) Create(context.Context, orgcap.DeptCreateInput) (int, error) {
	return 0, nil
}

// Update is unused by host-call demo tests.
func (s fakeOrgDepartmentService) Update(context.Context, orgcap.DeptUpdateInput) error {
	return nil
}

// Delete is unused by host-call demo tests.
func (s fakeOrgDepartmentService) Delete(context.Context, int) error {
	return nil
}

// Get returns no fake post projection.
func (s fakeOrgPostService) Get(context.Context, int) (*orgcap.PostInfo, error) {
	return nil, nil
}

// BatchGet returns no fake post projections.
func (s fakeOrgPostService) BatchGet(context.Context, []int) (*capmodel.BatchResult[*orgcap.PostInfo, int], error) {
	return &capmodel.BatchResult[*orgcap.PostInfo, int]{Items: map[int]*orgcap.PostInfo{}}, nil
}

// List returns no fake post projections.
func (s fakeOrgPostService) List(context.Context, orgcap.PostListInput) (*capmodel.PageResult[*orgcap.PostInfo], error) {
	return &capmodel.PageResult[*orgcap.PostInfo]{Items: []*orgcap.PostInfo{}}, nil
}

// ListOptions returns no fake post options.
func (s fakeOrgPostService) ListOptions(context.Context, orgcap.PostOptionsInput) (*capmodel.PageResult[*orgcap.PostOption], error) {
	return &capmodel.PageResult[*orgcap.PostOption]{Items: []*orgcap.PostOption{}}, nil
}

// EnsureVisible accepts fake post identifiers.
func (s fakeOrgPostService) EnsureVisible(context.Context, []int) error {
	return nil
}

// Create is unused by host-call demo tests.
func (s fakeOrgPostService) Create(context.Context, orgcap.PostCreateInput) (int, error) {
	return 0, nil
}

// Update is unused by host-call demo tests.
func (s fakeOrgPostService) Update(context.Context, orgcap.PostUpdateInput) error {
	return nil
}

// Delete is unused by host-call demo tests.
func (s fakeOrgPostService) Delete(context.Context, int) error {
	return nil
}

// fakeTenantHostService returns deterministic tenant capability values for
// unit tests.
type fakeTenantHostService struct{}

// Status returns a deterministic tenant capability status.
func (s *fakeTenantHostService) Status(_ context.Context) capmodel.CapabilityStatus {
	return capmodel.CapabilityStatus{
		CapabilityID:   tenantcap.CapabilityTenantV1,
		Available:      true,
		ActiveProvider: tenantcap.ProviderPluginID,
	}
}

// Available reports that the fake tenant capability is active.
func (s *fakeTenantHostService) Available(_ context.Context) bool {
	return true
}

// Context returns the fake tenant context subresource.
func (s *fakeTenantHostService) Context() tenantcap.ContextService {
	return s
}

// Directory returns the fake tenant directory subresource.
func (s *fakeTenantHostService) Directory() tenantcap.DirectoryService {
	return s
}

// Membership returns the fake tenant membership subresource.
func (s *fakeTenantHostService) Membership() tenantcap.MembershipService {
	return s
}

// Plugins returns no tenant-plugin governance subresource for this demo fake.
func (s *fakeTenantHostService) Plugins() tenantcap.PluginService {
	return nil
}

// Filter returns no tenant filter context subresource for this demo fake.
func (s *fakeTenantHostService) Filter() tenantcap.FilterService {
	return nil
}

// Current returns one deterministic current tenant.
func (s *fakeTenantHostService) Current(_ context.Context) tenantcap.TenantID {
	return tenantcap.TenantID(7)
}

// Info returns one deterministic current tenant projection.
func (s *fakeTenantHostService) Info(context.Context) (*tenantcap.TenantInfo, error) {
	return &tenantcap.TenantInfo{ID: tenantcap.TenantID(7), Code: "tenant-demo", Name: "Tenant Demo", Status: "active"}, nil
}

// PlatformBypass reports that the fake request uses tenant filtering.
func (s *fakeTenantHostService) PlatformBypass(_ context.Context) bool {
	return false
}

// EnsureVisible accepts the deterministic current tenant.
func (s *fakeTenantHostService) EnsureVisible(_ context.Context, _ []tenantcap.TenantID) error {
	return nil
}

// Get returns one deterministic tenant projection.
func (s *fakeTenantHostService) Get(context.Context, tenantcap.TenantID) (*tenantcap.TenantInfo, error) {
	return s.Info(context.Background())
}

// BatchGet returns deterministic tenant projections.
func (s *fakeTenantHostService) BatchGet(_ context.Context, tenantIDs []tenantcap.TenantID) (*capmodel.BatchResult[*tenantcap.TenantInfo, tenantcap.TenantID], error) {
	result := &capmodel.BatchResult[*tenantcap.TenantInfo, tenantcap.TenantID]{
		Items:      make(map[tenantcap.TenantID]*tenantcap.TenantInfo, len(tenantIDs)),
		MissingIDs: []tenantcap.TenantID{},
	}
	for _, tenantID := range tenantIDs {
		result.Items[tenantID] = &tenantcap.TenantInfo{ID: tenantID, Code: "tenant-demo", Name: "Tenant Demo", Status: "active"}
	}
	return result, nil
}

// List returns deterministic tenant projections.
func (s *fakeTenantHostService) List(context.Context, tenantcap.ListInput) (*capmodel.PageResult[*tenantcap.TenantInfo], error) {
	return &capmodel.PageResult[*tenantcap.TenantInfo]{
		Items: []*tenantcap.TenantInfo{{ID: tenantcap.TenantID(7), Code: "tenant-demo", Name: "Tenant Demo", Status: "active"}},
		Total: 1,
	}, nil
}

// ListByUser returns deterministic current-user tenants.
func (s *fakeTenantHostService) ListByUser(_ context.Context, _ int) ([]tenantcap.TenantInfo, error) {
	return []tenantcap.TenantInfo{{
		ID:     tenantcap.TenantID(7),
		Code:   "tenant-demo",
		Name:   "Tenant Demo",
		Status: "active",
	}}, nil
}

// Validate accepts deterministic user tenant memberships.
func (s *fakeTenantHostService) Validate(context.Context, int, tenantcap.TenantID) error {
	return nil
}

// TestRunHostCallDemoConfigReadsPluginAndHostConfigValues verifies the dynamic
// demo reads plugin config and public host config values through separate
// governed clients.
func TestRunHostCallDemoConfigReadsPluginAndHostConfigValues(t *testing.T) {
	service := &serviceImpl{
		pluginsSvc: &fakePluginHostService{
			config: &fakePluginConfigService{
				strings: map[string]configStringResult{
					hostCallDemoPluginGreetingKey: {
						value: "Hello from test config",
						found: true,
					},
				},
				bools: map[string]configBoolResult{
					hostCallDemoPluginFeatureKey: {
						value: true,
						found: true,
					},
				},
			},
		},
		hostConfigSvc: &fakeHostConfigHostService{
			strings: map[string]configStringResult{
				hostCallDemoWorkspaceKey: {
					value: "/tmp/linapro",
					found: true,
				},
				hostCallDemoI18nDefaultKey: {
					value: "zh-CN",
					found: true,
				},
			},
			bools: map[string]configBoolResult{
				hostCallDemoI18nEnabledKey: {
					value: true,
					found: true,
				},
			},
		},
	}

	payload, err := service.runHostCallDemoConfig(t.Context())
	if err != nil {
		t.Fatalf("expected config demo to succeed, got error: %v", err)
	}
	if !payload.Plugin.GreetingFound || payload.Plugin.Greeting != "Hello from test config" {
		t.Fatalf("unexpected plugin greeting payload: %#v", payload.Plugin)
	}
	if !payload.Plugin.FeatureEnabledFound || !payload.Plugin.FeatureEnabled {
		t.Fatalf("unexpected plugin feature payload: %#v", payload.Plugin)
	}
	if !payload.HostConfig.WorkspaceBasePathFound || payload.HostConfig.WorkspaceBasePath != "/tmp/linapro" {
		t.Fatalf("unexpected host workspace payload: %#v", payload.HostConfig)
	}
	if !payload.HostConfig.I18nDefaultFound || payload.HostConfig.I18nDefault != "zh-CN" {
		t.Fatalf("unexpected host i18n default payload: %#v", payload.HostConfig)
	}
	if !payload.HostConfig.I18nEnabledFound || !payload.HostConfig.I18nEnabled {
		t.Fatalf("unexpected host i18n enabled payload: %#v", payload.HostConfig)
	}
}

// TestRunHostCallDemoManifestReadsAuthorizedResources verifies the dynamic
// demo reads only the manifest resources declared for the manifest host
// service example.
func TestRunHostCallDemoManifestReadsAuthorizedResources(t *testing.T) {
	service := &serviceImpl{
		manifestSvc: &fakeManifestHostService{
			texts: map[string]manifestTextResult{
				hostCallDemoManifestProfilePath: {
					value: "name: demo-dynamic-profile\ntier: sample\nowner: linapro\n",
					found: true,
				},
				hostCallDemoManifestConfigPath: {
					value: "demo:\n  greeting: Hello from test manifest config\n  featureEnabled: true\n",
					found: true,
				},
			},
			profile: hostCallDemoManifestProfile{
				Name:  "demo-dynamic-profile",
				Tier:  "sample",
				Owner: "linapro",
			},
		},
	}

	payload, err := service.runHostCallDemoManifest(t.Context())
	if err != nil {
		t.Fatalf("expected manifest demo to succeed, got error: %v", err)
	}
	if payload.ProfilePath != hostCallDemoManifestProfilePath || !payload.ProfileFound {
		t.Fatalf("unexpected profile path/found payload: %#v", payload)
	}
	if payload.ProfileName != "demo-dynamic-profile" ||
		payload.ProfileTier != "sample" ||
		payload.ProfileOwner != "linapro" {
		t.Fatalf("unexpected profile payload: %#v", payload)
	}
	if payload.ConfigPath != hostCallDemoManifestConfigPath || !payload.ConfigFound {
		t.Fatalf("unexpected config path/found payload: %#v", payload)
	}
	if !strings.Contains(payload.ConfigBodyPreview, "Hello from test manifest config") {
		t.Fatalf("unexpected config preview payload: %#v", payload)
	}
	if payload.BatchReadCount != 2 || payload.MissingPathCount != 0 || payload.ListedCount != 2 {
		t.Fatalf("unexpected manifest batch payload: %#v", payload)
	}
}

// TestRunHostCallDemoOrgTenantReadsCapabilityServices verifies the dynamic demo
// exercises organization and tenant host services through dedicated clients.
func TestRunHostCallDemoOrgTenantReadsCapabilityServices(t *testing.T) {
	service := &serviceImpl{
		orgSvc:    &fakeOrgHostService{},
		tenantSvc: &fakeTenantHostService{},
	}
	input := &HostCallDemoInput{UserID: 42}

	orgPayload, err := service.runHostCallDemoOrg(context.Background(), input)
	if err != nil {
		t.Fatalf("expected org demo to succeed, got error: %v", err)
	}
	if !orgPayload.Available || orgPayload.CapabilityID != orgcap.CapabilityOrgV1 {
		t.Fatalf("unexpected org status payload: %#v", orgPayload)
	}
	if orgPayload.AssignmentCount != 1 ||
		orgPayload.CurrentUserDeptCount != 1 ||
		orgPayload.CurrentUserPostCount != 2 {
		t.Fatalf("unexpected org projection payload: %#v", orgPayload)
	}

	tenantPayload, err := service.runHostCallDemoTenant(context.Background(), input)
	if err != nil {
		t.Fatalf("expected tenant demo to succeed, got error: %v", err)
	}
	if !tenantPayload.Available || tenantPayload.CapabilityID != tenantcap.CapabilityTenantV1 {
		t.Fatalf("unexpected tenant status payload: %#v", tenantPayload)
	}
	if tenantPayload.CurrentTenantID != 7 ||
		tenantPayload.PlatformBypass ||
		tenantPayload.UserTenantCount != 1 ||
		!tenantPayload.Visible {
		t.Fatalf("unexpected tenant projection payload: %#v", tenantPayload)
	}
}

// TestRunHostCallDemoAIReadsOwnerMethodStatus verifies the dynamic demo uses
// the linapro-ai-core owner contract for AI method status reads.
func TestRunHostCallDemoAIReadsOwnerMethodStatus(t *testing.T) {
	service := &serviceImpl{
		aiSvc: aicap.New(aitext.NewUnavailable()),
	}

	payload, err := service.runHostCallDemoAI(context.Background())
	if err != nil {
		t.Fatalf("expected AI demo to succeed, got error: %v", err)
	}
	if payload.Owner != spi.OwnerPluginID || payload.Service != spi.ServiceAI || payload.Version != spi.VersionV1 {
		t.Fatalf("unexpected owner identity payload: %#v", payload)
	}
	if payload.CapabilityType != string(aicommon.CapabilityTypeText) ||
		payload.CapabilityMethod != string(aicommon.CapabilityMethodTextGenerate) {
		t.Fatalf("unexpected AI method payload: %#v", payload)
	}
	if payload.Available || payload.CapabilityID != aitext.CapabilityAITextV1 || payload.Reason == "" {
		t.Fatalf("unexpected AI method status payload: %#v", payload)
	}
}

// TestRunHostCallDemoBizCtxCacheLockReadsCapabilityServices verifies the
// dynamic demo exercises bizctx, cache, and lock host services through
// dedicated clients.
func TestRunHostCallDemoBizCtxCacheLockReadsCapabilityServices(t *testing.T) {
	service := &serviceImpl{
		bizCtxSvc: &fakeBizCtxHostService{current: bizctxcap.CurrentContext{
			UserID:         7,
			Username:       "demo-user",
			TenantID:       3,
			Permissions:    []string{"plugin:view", "plugin:edit", "plugin:cache"},
			IsSuperAdmin:   false,
			PlatformBypass: true,
			ActingAsTenant: true,
		}},
		cacheSvc: &fakeCacheHostService{},
		lockSvc:  &fakeLockHostService{},
	}

	bizCtxPayload, err := service.runHostCallDemoBizCtx(context.Background())
	if err != nil {
		t.Fatalf("expected bizctx demo to succeed, got error: %v", err)
	}
	if bizCtxPayload.PermissionCount != 3 || !bizCtxPayload.PlatformBypass || !bizCtxPayload.ActingAsTenant {
		t.Fatalf("unexpected bizctx payload: %#v", bizCtxPayload)
	}

	cachePayload, err := service.runHostCallDemoCache(context.Background(), "demo-key")
	if err != nil {
		t.Fatalf("expected cache demo to succeed, got error: %v", err)
	}
	if cachePayload.Namespace != hostCallDemoCacheNamespace || !cachePayload.SingleFound || cachePayload.BatchSetCount != 2 || cachePayload.BatchReadCount != 3 || cachePayload.MissingCount != 1 || cachePayload.IncrementedValue != 2 || !cachePayload.ExpireUpdated || !cachePayload.Deleted {
		t.Fatalf("unexpected cache payload: %#v", cachePayload)
	}

	lockPayload, err := service.runHostCallDemoLock(context.Background())
	if err != nil {
		t.Fatalf("expected lock demo to succeed, got error: %v", err)
	}
	if lockPayload.Name != hostCallDemoLockName || !lockPayload.Acquired || !lockPayload.Renewed || !lockPayload.Released || !lockPayload.TicketIssued {
		t.Fatalf("unexpected lock payload: %#v", lockPayload)
	}
}

// TestRunHostCallDemoStorageUsesBatchAndCursorCalls verifies the dynamic demo
// exercises batch stat, cursor list, and batch delete storage host services.
func TestRunHostCallDemoStorageUsesBatchAndCursorCalls(t *testing.T) {
	service := &serviceImpl{
		storageSvc: &fakeStorageHostService{},
	}

	payload, err := service.runHostCallDemoStorage(context.Background(), "demo-plugin", "demo-key")
	if err != nil {
		t.Fatalf("expected storage demo to succeed, got error: %v", err)
	}
	if payload.ListedCount != 1 || payload.CursorListedCount != 1 || payload.BatchStatCount != 1 || payload.BatchStatMissingCount != 0 || !payload.BatchDeleted || !payload.Deleted {
		t.Fatalf("unexpected storage payload: %#v", payload)
	}
}
