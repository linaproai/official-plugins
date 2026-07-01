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

	operTypeMap := service.buildStringDictLabelMap(context.Background(), dictTypeOperType)
	statusMap := service.buildIntDictLabelMap(context.Background(), dictTypeOperStatus)

	if actual := service.exportOperTypeText(context.Background(), operlogtype.OperTypeExport.String(), operTypeMap); actual != "Domain Export" {
		t.Fatalf("expected dictcap operation type label, got %q", actual)
	}
	if actual := service.exportStatusText(context.Background(), OperStatusSuccess, statusMap); actual != "Domain Success" {
		t.Fatalf("expected dictcap status label, got %q", actual)
	}
	if dict.lastType != dictcap.Type(dictTypeOperStatus) {
		t.Fatalf("expected dictcap type=%s, got type=%s", dictTypeOperStatus, dict.lastType)
	}
}
