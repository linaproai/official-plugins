// This file verifies operation-log runtime i18n projections.

package operlog

import (
	"context"
	"testing"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/dictcap"
	"lina-plugin-linapro-monitor-operlog/backend/internal/model/operlogtype"
)

// fakeI18nService provides deterministic runtime translations for operation-log tests.
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

// ListValues returns configured labels as one deterministic page.
func (s *fakeDictService) ListValues(_ context.Context, capCtx capmodel.CapabilityContext, input dictcap.ListValuesInput) (*capmodel.PageResult[*dictcap.LabelProjection], error) {
	s.lastPluginID = capCtx.PluginID
	s.lastType = input.Type
	result := &capmodel.PageResult[*dictcap.LabelProjection]{Items: []*dictcap.LabelProjection{}}
	for value, label := range s.labels {
		result.Items = append(result.Items, &dictcap.LabelProjection{
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
func (s *fakeDictService) EnsureValuesVisible(_ context.Context, capCtx capmodel.CapabilityContext, input dictcap.ResolveInput) error {
	s.lastPluginID = capCtx.PluginID
	s.lastType = input.Type
	return nil
}

// TestExportHeadersUseRuntimeI18N verifies operation-log export headers resolve
// through runtime i18n keys.
func TestExportHeadersUseRuntimeI18N(t *testing.T) {
	service := &serviceImpl{i18nSvc: fakeI18nService{messages: map[string]string{
		"plugin.linapro-monitor-operlog.fields.moduleName":     "模块名称",
		"plugin.linapro-monitor-operlog.fields.operSummary":    "操作摘要",
		"plugin.linapro-monitor-operlog.fields.operType":       "操作类型",
		"plugin.linapro-monitor-operlog.fields.operator":       "操作人员",
		"plugin.linapro-monitor-operlog.fields.requestMethod":  "请求方式",
		"plugin.linapro-monitor-operlog.fields.requestUrl":     "请求地址",
		"plugin.linapro-monitor-operlog.fields.ipAddress":      "IP地址",
		"plugin.linapro-monitor-operlog.fields.requestParams":  "请求参数",
		"plugin.linapro-monitor-operlog.fields.responseResult": "返回结果",
		"plugin.linapro-monitor-operlog.fields.operResult":     "操作结果",
		"plugin.linapro-monitor-operlog.fields.errorInfo":      "错误信息",
		"plugin.linapro-monitor-operlog.fields.durationMs":     "耗时(毫秒)",
		"plugin.linapro-monitor-operlog.fields.operTime":       "操作时间",
	}}}

	actual := service.exportHeaders(context.Background())
	expected := []string{
		"模块名称", "操作摘要", "操作类型", "操作人员", "请求方式", "请求地址", "IP地址",
		"请求参数", "返回结果", "操作结果", "错误信息", "耗时(毫秒)", "操作时间",
	}
	for index, item := range expected {
		if actual[index] != item {
			t.Fatalf("expected header %d to be %q, got %q", index, item, actual[index])
		}
	}
}

// TestExportTypeAndStatusTextUseRuntimeI18N verifies operation type and result
// export text resolves through dictionary runtime i18n keys.
func TestExportTypeAndStatusTextUseRuntimeI18N(t *testing.T) {
	service := &serviceImpl{i18nSvc: fakeI18nService{messages: map[string]string{
		"dict.sys_oper_type.export.label": "导出",
		"dict.sys_oper_status.0.label":    "成功",
		"dict.sys_oper_status.1.label":    "失败",
	}}}

	if actual := service.exportOperTypeText(context.Background(), operlogtype.OperTypeExport.String(), nil); actual != "导出" {
		t.Fatalf("expected export type label, got %q", actual)
	}
	if actual := service.exportStatusText(context.Background(), OperStatusSuccess, nil); actual != "成功" {
		t.Fatalf("expected success status label, got %q", actual)
	}
	if actual := service.exportStatusText(context.Background(), OperStatusFail, nil); actual != "失败" {
		t.Fatalf("expected failure status label, got %q", actual)
	}
}

// TestExportTypeAndStatusTextUseDictCapability verifies backend export labels
// are resolved through dictcap instead of a plugin-generated host dictionary DAO.
func TestExportTypeAndStatusTextUseDictCapability(t *testing.T) {
	dict := &fakeDictService{labels: map[dictcap.Value]string{
		dictcap.Value(operlogtype.OperTypeExport.String()): "Domain Export",
		dictcap.Value("0"): "Domain Success",
	}}
	service := &serviceImpl{dictSvc: dict}

	operTypeMap := service.buildStringDictLabelMap(context.Background(), DictTypeOperType)
	statusMap := service.buildIntDictLabelMap(context.Background(), DictTypeOperStatus)

	if actual := service.exportOperTypeText(context.Background(), operlogtype.OperTypeExport.String(), operTypeMap); actual != "Domain Export" {
		t.Fatalf("expected dictcap operation type label, got %q", actual)
	}
	if actual := service.exportStatusText(context.Background(), OperStatusSuccess, statusMap); actual != "Domain Success" {
		t.Fatalf("expected dictcap status label, got %q", actual)
	}
	if dict.lastPluginID != pluginID || dict.lastType != dictcap.Type(DictTypeOperStatus) {
		t.Fatalf("expected dictcap context plugin=%s type=%s, got plugin=%s type=%s", pluginID, DictTypeOperStatus, dict.lastPluginID, dict.lastType)
	}
}
