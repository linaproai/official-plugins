// This file verifies login-log message localization behavior.

package loginlog

import (
	"context"
	"testing"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/dictcap"
	"lina-core/pkg/plugin/pluginhost"
)

// fakeI18nService provides deterministic runtime translations for unit tests.
type fakeI18nService struct {
	messages map[string]string
}

// GetLocale returns the fixed test locale.
func (s fakeI18nService) GetLocale(_ context.Context) string {
	return "zh-CN"
}

// Translate resolves known test keys and otherwise returns the fallback text.
func (s fakeI18nService) Translate(_ context.Context, key string, fallback string) string {
	if value, ok := s.messages[key]; ok {
		return value
	}
	return fallback
}

// fakeDictService returns deterministic dictionary-domain labels.
type fakeDictService struct {
	labels   map[dictcap.Value]string
	lastType dictcap.Type
}

// Type is unused by these tests.
func (s *fakeDictService) Type() dictcap.TypeService {
	return nil
}

// Value returns the fake value service.
func (s *fakeDictService) Value() dictcap.ValueService {
	return s
}

// Refresh is unused by these tests.
func (s *fakeDictService) Refresh(context.Context, dictcap.Type) error {
	return nil
}

// Get is unused by these tests.
func (s *fakeDictService) Get(context.Context, int) (*dictcap.ValueInfo, error) {
	return nil, nil
}

// BatchGet is unused by these tests.
func (s *fakeDictService) BatchGet(context.Context, dictcap.BatchGetValuesInput) (*capmodel.BatchResult[*dictcap.ValueInfo, dictcap.Value], error) {
	return &capmodel.BatchResult[*dictcap.ValueInfo, dictcap.Value]{Items: map[dictcap.Value]*dictcap.ValueInfo{}}, nil
}

// ResolveLabels returns configured labels using dictcap batch semantics.
func (s *fakeDictService) ResolveLabels(ctx context.Context, input dictcap.ResolveInput) (*capmodel.BatchResult[*dictcap.LabelInfo, dictcap.Value], error) {
	s.lastType = input.Type
	result := &capmodel.BatchResult[*dictcap.LabelInfo, dictcap.Value]{
		Items:      map[dictcap.Value]*dictcap.LabelInfo{},
		MissingIDs: []dictcap.Value{},
	}
	for _, value := range input.Values {
		label, ok := s.labels[value]
		if !ok {
			result.MissingIDs = append(result.MissingIDs, value)
			continue
		}
		result.Items[value] = &dictcap.LabelInfo{
			Type:     input.Type,
			Value:    value,
			LabelKey: "dict." + string(input.Type) + "." + string(value) + ".label",
			Label:    label,
		}
	}
	return result, nil
}

// List returns configured labels as one deterministic page.
func (s *fakeDictService) List(ctx context.Context, input dictcap.ListValuesInput) (*capmodel.PageResult[*dictcap.ValueInfo], error) {
	s.lastType = input.Type
	result := &capmodel.PageResult[*dictcap.ValueInfo]{Items: []*dictcap.ValueInfo{}}
	for value, label := range s.labels {
		result.Items = append(result.Items, &dictcap.ValueInfo{
			Type:     input.Type,
			Value:    value,
			LabelKey: "dict." + string(input.Type) + "." + string(value) + ".label",
			Label:    label,
		})
	}
	result.Total = len(result.Items)
	return result, nil
}

// EnsureValuesVisible accepts dictionary values used by localization tests.
func (s *fakeDictService) EnsureValuesVisible(ctx context.Context, input dictcap.ResolveInput) error {
	s.lastType = input.Type
	return nil
}

// EnsureVisible is unused by these tests.
func (s *fakeDictService) EnsureVisible(context.Context, []int) error {
	return nil
}

// Create is unused by these tests.
func (s *fakeDictService) Create(context.Context, dictcap.CreateValueInput) (int, error) {
	return 0, nil
}

// Update is unused by these tests.
func (s *fakeDictService) Update(context.Context, dictcap.UpdateValueInput) error {
	return nil
}

// Delete is unused by these tests.
func (s *fakeDictService) Delete(context.Context, int) error {
	return nil
}

// DeleteByType is unused by these tests.
func (s *fakeDictService) DeleteByType(context.Context, dictcap.Type) error {
	return nil
}

// TestTranslateLoginLogMessageResolvesStableReason verifies that login-log
// display messages are translated from stable auth lifecycle reason codes.
func TestTranslateLoginLogMessageResolvesStableReason(t *testing.T) {
	service := &serviceImpl{i18nSvc: fakeI18nService{messages: map[string]string{
		loginLogMessagePrefix + ".loginSuccessful": "登录成功",
	}}}

	actual := service.translateLoginLogMessage(context.Background(), pluginhost.AuthHookReasonLoginSuccessful)
	if actual != "登录成功" {
		t.Fatalf("expected stable reason to resolve, got %q", actual)
	}
}

// TestTranslateLoginLogMessagePreservesRawMessages verifies that custom raw
// audit messages are not interpreted through legacy text-to-key mappings.
func TestTranslateLoginLogMessagePreservesRawMessages(t *testing.T) {
	service := &serviceImpl{i18nSvc: fakeI18nService{messages: map[string]string{
		"plugin.linapro-monitor-loginlog.logMessage.loginSuccessful": "登录成功",
	}}}

	actual := service.translateLoginLogMessage(context.Background(), "Login successful")
	if actual != "Login successful" {
		t.Fatalf("expected raw message to remain unchanged, got %q", actual)
	}
}

// TestExportHeadersUseRuntimeI18N verifies login-log export headers resolve
// through runtime i18n keys.
func TestExportHeadersUseRuntimeI18N(t *testing.T) {
	service := &serviceImpl{i18nSvc: fakeI18nService{messages: map[string]string{
		"plugin.linapro-monitor-loginlog.fields.userName":  "用户账号",
		"plugin.linapro-monitor-loginlog.fields.status":    "登录状态",
		"plugin.linapro-monitor-loginlog.fields.ipAddress": "IP地址",
		"plugin.linapro-monitor-loginlog.fields.browser":   "浏览器",
		"plugin.linapro-monitor-loginlog.fields.os":        "操作系统",
		"plugin.linapro-monitor-loginlog.fields.message":   "消息",
		"plugin.linapro-monitor-loginlog.fields.loginTime": "登录时间",
	}}}

	actual := service.exportHeaders(context.Background())
	expected := []string{"用户账号", "登录状态", "IP地址", "浏览器", "操作系统", "消息", "登录时间"}
	for index, item := range expected {
		if actual[index] != item {
			t.Fatalf("expected header %d to be %q, got %q", index, item, actual[index])
		}
	}
}

// TestExportStatusTextUseRuntimeI18N verifies fallback login status labels are
// still resolved through dictionary runtime i18n keys.
func TestExportStatusTextUseRuntimeI18N(t *testing.T) {
	service := &serviceImpl{i18nSvc: fakeI18nService{messages: map[string]string{
		"dict.sys_login_status.0.label": "成功",
		"dict.sys_login_status.1.label": "失败",
	}}}

	if actual := service.exportStatusText(context.Background(), loginStatusSuccess, nil); actual != "成功" {
		t.Fatalf("expected success label, got %q", actual)
	}
	if actual := service.exportStatusText(context.Background(), loginStatusFail, nil); actual != "失败" {
		t.Fatalf("expected failed label, got %q", actual)
	}
}

// TestExportStatusTextUseDictCapability verifies backend export status labels
// are resolved through dictcap instead of a plugin-generated host dictionary DAO.
func TestExportStatusTextUseDictCapability(t *testing.T) {
	dict := &fakeDictService{labels: map[dictcap.Value]string{
		dictcap.Value("0"): "Domain Success",
	}}
	service := &serviceImpl{dictSvc: dict}

	statusMap := service.buildIntDictLabelMap(context.Background(), dictTypeLoginStatus)

	if actual := service.exportStatusText(context.Background(), loginStatusSuccess, statusMap); actual != "Domain Success" {
		t.Fatalf("expected dictcap login status label, got %q", actual)
	}
	if dict.lastType != dictcap.Type(dictTypeLoginStatus) {
		t.Fatalf("expected dictcap type=%s, got type=%s", dictTypeLoginStatus, dict.lastType)
	}
}
