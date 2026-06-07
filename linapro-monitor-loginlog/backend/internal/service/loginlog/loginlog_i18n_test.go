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

// FindMessageKeys is unused by these tests and returns no matches.
func (s fakeI18nService) FindMessageKeys(_ context.Context, _ string, _ string) []string {
	return []string{}
}

// fakeDictService returns deterministic dictionary-domain labels.
type fakeDictService struct {
	labels       map[dictcap.Value]string
	lastPluginID string
	lastType     dictcap.Type
}

// ResolveLabels returns configured labels using dictcap batch semantics.
func (s *fakeDictService) ResolveLabels(_ context.Context, capCtx capmodel.CapabilityContext, input dictcap.ResolveInput) (*capmodel.BatchResult[*dictcap.LabelProjection, dictcap.Value], error) {
	s.lastPluginID = capCtx.PluginID
	s.lastType = input.Type
	result := &capmodel.BatchResult[*dictcap.LabelProjection, dictcap.Value]{
		Items:      map[dictcap.Value]*dictcap.LabelProjection{},
		MissingIDs: []dictcap.Value{},
	}
	for _, value := range input.Values {
		label, ok := s.labels[value]
		if !ok {
			result.MissingIDs = append(result.MissingIDs, value)
			continue
		}
		result.Items[value] = &dictcap.LabelProjection{
			Type:     input.Type,
			Value:    value,
			LabelKey: "dict." + string(input.Type) + "." + string(value) + ".label",
			Label:    label,
		}
	}
	return result, nil
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

	if actual := service.exportStatusText(context.Background(), LoginStatusSuccess, nil); actual != "成功" {
		t.Fatalf("expected success label, got %q", actual)
	}
	if actual := service.exportStatusText(context.Background(), LoginStatusFail, nil); actual != "失败" {
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

	statusMap := service.buildIntDictLabelMap(context.Background(), DictTypeLoginStatus)

	if actual := service.exportStatusText(context.Background(), LoginStatusSuccess, statusMap); actual != "Domain Success" {
		t.Fatalf("expected dictcap login status label, got %q", actual)
	}
	if dict.lastPluginID != pluginID || dict.lastType != dictcap.Type(DictTypeLoginStatus) {
		t.Fatalf("expected dictcap context plugin=%s type=%s, got plugin=%s type=%s", pluginID, DictTypeLoginStatus, dict.lastPluginID, dict.lastType)
	}
}
