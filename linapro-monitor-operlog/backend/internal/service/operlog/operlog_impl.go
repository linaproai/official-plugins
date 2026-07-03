// operlog_impl.go implements operation-log persistence, tenant-filtered
// queries, cleanup, and Excel export for the linapro-monitor-operlog plugin. It applies
// host tenant filtering before data access and resolves dictionary/i18n labels
// without depending on host-internal service packages.

package operlog

import (
	"bytes"
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/xuri/excelize/v2"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/excelutil"
	"lina-core/pkg/gdbutil"
	"lina-core/pkg/plugin/capability/apidoccap"
	"lina-core/pkg/plugin/capability/dictcap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"lina-plugin-linapro-monitor-operlog/backend/internal/dao"
	"lina-plugin-linapro-monitor-operlog/backend/internal/model/do"
	"lina-plugin-linapro-monitor-operlog/backend/internal/model/operlogtype"
)

// Create inserts one operation-log record.
func (s *serviceImpl) Create(ctx context.Context, in CreateInput) error {
	operType := in.OperType
	if !operlogtype.IsSupported(operType) {
		operType = operlogtype.OperTypeOther
	}
	auditContext := resolveAuditTenantContext(
		ctx,
		s.pluginTableFilter(),
		in.TenantID,
		in.ActingUserID,
		in.OnBehalfOfTenantID,
		in.IsImpersonation,
	)

	_, err := dao.Operlog.Ctx(ctx).Data(do.Operlog{
		TenantId:           auditContext.TenantID,
		ActingUserId:       auditContext.ActingUserID,
		OnBehalfOfTenantId: auditContext.OnBehalfOfTenantID,
		IsImpersonation:    auditContext.IsImpersonation,
		Title:              in.Title,
		OperSummary:        in.OperSummary,
		RouteOwner:         in.RouteOwner,
		RouteMethod:        in.RouteMethod,
		RoutePath:          in.RoutePath,
		RouteDocKey:        in.RouteDocKey,
		OperType:           operType.String(),
		Method:             in.Method,
		RequestMethod:      in.RequestMethod,
		OperName:           in.OperName,
		OperUrl:            in.OperUrl,
		OperIp:             in.OperIp,
		OperParam:          in.OperParam,
		JsonResult:         in.JsonResult,
		Status:             in.Status,
		ErrorMsg:           in.ErrorMsg,
		CostTime:           in.CostTime,
		OperTime:           timePtr(time.Now()),
	}).Insert()
	return err
}

// timePtr returns a pointer to value for generated DO time fields that preserve
// database NULL semantics with *time.Time.
func timePtr(value time.Time) *time.Time {
	return &value
}

// List queries the paginated operation-log list.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	model := tenantspi.ApplyPluginTableFilter(ctx, s.pluginTableFilter(), dao.Operlog.Ctx(ctx), "")
	titleOperationKeys := s.findLocalizedRouteTitleOperationKeys(ctx, in.Title)
	model = applyOperLogFilters(model, in.Title, titleOperationKeys, in.OperName, in.OperType, in.Status, in.BeginTime, in.EndTime)

	total, err := model.Count()
	if err != nil {
		return nil, err
	}

	allowedSortFields := map[string]string{
		"id":        colID,
		"operTime":  colOperTime,
		"oper_time": colOperTime,
		"costTime":  colCostTime,
		"cost_time": colCostTime,
	}
	orderBy := colOperTime
	if field, ok := allowedSortFields[in.OrderBy]; ok {
		orderBy = field
	}
	direction := gdbutil.NormalizeOrderDirectionOrDefault(in.OrderDirection, gdbutil.OrderDirectionDESC)

	list := make([]*OperLogEntity, 0)
	err = gdbutil.ApplyModelOrder(
		model.Page(in.PageNum, in.PageSize),
		orderBy,
		direction,
	).Scan(&list)
	if err != nil {
		return nil, err
	}
	s.localizeRecords(ctx, list)

	return &ListOutput{List: list, Total: total}, nil
}

// GetById retrieves one operation-log record by primary key.
func (s *serviceImpl) GetById(ctx context.Context, id int) (*OperLogEntity, error) {
	var record *OperLogEntity
	err := tenantspi.ApplyPluginTableFilter(ctx, s.pluginTableFilter(), dao.Operlog.Ctx(ctx), "").Where(colID, id).Scan(&record)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, bizerr.NewCode(CodeOperLogNotFound)
	}
	s.localizeRecord(ctx, record)
	return record, nil
}

// Clean hard-deletes operation logs within one optional time range.
func (s *serviceImpl) Clean(ctx context.Context, in CleanInput) (int, error) {
	model := tenantspi.ApplyPluginTableFilter(ctx, s.pluginTableFilter(), dao.Operlog.Ctx(ctx), "")
	hasFilter := false
	if in.BeginTime != "" {
		model = model.WhereGTE(colOperTime, in.BeginTime)
		hasFilter = true
	}
	if in.EndTime != "" {
		model = model.WhereLTE(colOperTime, normalizeEndTime(in.EndTime))
		hasFilter = true
	}
	if !hasFilter {
		model = model.Where("1 = 1")
	}

	result, err := model.Delete()
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(affected), nil
}

// DeleteByIds hard-deletes operation logs by ID list.
func (s *serviceImpl) DeleteByIds(ctx context.Context, ids []int) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result, err := tenantspi.ApplyPluginTableFilter(ctx, s.pluginTableFilter(), dao.Operlog.Ctx(ctx), "").WhereIn(colID, ids).Delete()
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(affected), nil
}

// Export generates an Excel workbook for operation logs.
func (s *serviceImpl) Export(ctx context.Context, in ExportInput) (data []byte, err error) {
	model := tenantspi.ApplyPluginTableFilter(ctx, s.pluginTableFilter(), dao.Operlog.Ctx(ctx), "")
	if len(in.Ids) > 0 {
		model = model.WhereIn(colID, in.Ids)
	} else {
		titleOperationKeys := s.findLocalizedRouteTitleOperationKeys(ctx, in.Title)
		model = applyOperLogFilters(model, in.Title, titleOperationKeys, in.OperName, in.OperType, in.Status, in.BeginTime, in.EndTime)
	}
	model = model.Limit(maxExportRows)

	allowedSortFields := map[string]string{
		"id":        colID,
		"operTime":  colOperTime,
		"oper_time": colOperTime,
		"costTime":  colCostTime,
		"cost_time": colCostTime,
	}
	orderBy := colOperTime
	if field, ok := allowedSortFields[in.OrderBy]; ok {
		orderBy = field
	}
	direction := gdbutil.NormalizeOrderDirectionOrDefault(in.OrderDirection, gdbutil.OrderDirectionDESC)

	list := make([]*OperLogEntity, 0)
	err = gdbutil.ApplyModelOrder(model, orderBy, direction).Scan(&list)
	if err != nil {
		return nil, err
	}
	s.localizeRecords(ctx, list)

	file := excelize.NewFile()
	defer excelutil.CloseFile(ctx, file, &err)
	sheet := "Sheet1"
	headers := s.exportHeaders(ctx)
	for index, header := range headers {
		if setErr := excelutil.SetCellValue(file, sheet, index+1, 1, header); setErr != nil {
			return nil, setErr
		}
	}

	operTypeMap := s.buildStringDictLabelMap(ctx, dictTypeOperType)
	statusMap := s.buildIntDictLabelMap(ctx, dictTypeOperStatus)
	for index, log := range list {
		row := index + 2
		if setErr := excelutil.SetCellValue(file, sheet, 1, row, log.Title); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValue(file, sheet, 2, row, log.OperSummary); setErr != nil {
			return nil, setErr
		}
		operTypeText := s.exportOperTypeText(ctx, log.OperType, operTypeMap)
		if setErr := excelutil.SetCellValue(file, sheet, 3, row, operTypeText); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValue(file, sheet, 4, row, log.OperName); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValue(file, sheet, 5, row, log.RequestMethod); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValue(file, sheet, 6, row, log.OperUrl); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValue(file, sheet, 7, row, log.OperIp); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValue(file, sheet, 8, row, log.OperParam); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValue(file, sheet, 9, row, log.JsonResult); setErr != nil {
			return nil, setErr
		}
		statusText := s.exportStatusText(ctx, log.Status, statusMap)
		if setErr := excelutil.SetCellValue(file, sheet, 10, row, statusText); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValue(file, sheet, 11, row, log.ErrorMsg); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValue(file, sheet, 12, row, log.CostTime); setErr != nil {
			return nil, setErr
		}
		if log.OperTime != nil {
			if setErr := excelutil.SetCellValue(file, sheet, 13, row, log.OperTime.String()); setErr != nil {
				return nil, setErr
			}
		}
	}

	var buffer bytes.Buffer
	if writeErr := file.Write(&buffer); writeErr != nil {
		return nil, writeErr
	}
	return buffer.Bytes(), nil
}

// exportHeaders returns localized Excel headers for operation-log export.
func (s *serviceImpl) exportHeaders(ctx context.Context) []string {
	headers := []exportHeader{
		{Key: "plugin.linapro-monitor-operlog.fields.moduleName", Fallback: "Module Name"},
		{Key: "plugin.linapro-monitor-operlog.fields.operSummary", Fallback: "Operation Summary"},
		{Key: "plugin.linapro-monitor-operlog.fields.operType", Fallback: "Operation Type"},
		{Key: "plugin.linapro-monitor-operlog.fields.operator", Fallback: "Operator"},
		{Key: "plugin.linapro-monitor-operlog.fields.requestMethod", Fallback: "Request Method"},
		{Key: "plugin.linapro-monitor-operlog.fields.requestUrl", Fallback: "Request URL"},
		{Key: "plugin.linapro-monitor-operlog.fields.ipAddress", Fallback: "IP Address"},
		{Key: "plugin.linapro-monitor-operlog.fields.requestParams", Fallback: "Request Parameters"},
		{Key: "plugin.linapro-monitor-operlog.fields.responseResult", Fallback: "Response Result"},
		{Key: "plugin.linapro-monitor-operlog.fields.operResult", Fallback: "Operation Result"},
		{Key: "plugin.linapro-monitor-operlog.fields.errorInfo", Fallback: "Error Information"},
		{Key: "plugin.linapro-monitor-operlog.fields.durationMs", Fallback: "Duration (ms)"},
		{Key: "plugin.linapro-monitor-operlog.fields.operTime", Fallback: "Operation Time"},
	}

	result := make([]string, 0, len(headers))
	for _, header := range headers {
		result = append(result, s.translate(ctx, header.Key, header.Fallback))
	}
	return result
}

// exportOperTypeText returns the localized export label for one operation type.
func (s *serviceImpl) exportOperTypeText(ctx context.Context, operType string, operTypeMap map[string]string) string {
	operTypeText, ok := operTypeMap[operType]
	if !ok {
		operTypeText = s.localizeDictValue(ctx, dictTypeOperType, operType, defaultOperTypeLabels[operlogtype.Normalize(operType)])
	}
	if operTypeText == "" {
		return operType
	}
	return operTypeText
}

// exportStatusText returns the localized export label for one operation status.
func (s *serviceImpl) exportStatusText(ctx context.Context, status int, statusMap map[int]string) string {
	statusText, ok := statusMap[status]
	if !ok {
		statusText = s.localizeDictValue(ctx, dictTypeOperStatus, strconv.Itoa(status), defaultOperStatusLabels[status])
	}
	return statusText
}

// localizeRecords translates route metadata fallback fields on every record.
func (s *serviceImpl) localizeRecords(ctx context.Context, records []*OperLogEntity) {
	if len(records) == 0 {
		return
	}
	if s == nil || s.apiDocSvc == nil {
		return
	}

	inputs := make([]apidoccap.RouteTextInput, 0, len(records))
	targets := make([]*OperLogEntity, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		inputs = append(inputs, apidoccap.RouteTextInput{
			OperationKey:    record.RouteDocKey,
			Method:          record.RouteMethod,
			Path:            record.RoutePath,
			FallbackTitle:   record.Title,
			FallbackSummary: record.OperSummary,
		})
		targets = append(targets, record)
	}
	if len(inputs) == 0 {
		return
	}

	outputs := s.apiDocSvc.ResolveRouteTexts(ctx, inputs)
	for index, output := range outputs {
		if index >= len(targets) {
			break
		}
		targets[index].Title = output.Title
		targets[index].OperSummary = output.Summary
	}
}

// localizeRecord translates route metadata fallback fields on one record.
func (s *serviceImpl) localizeRecord(ctx context.Context, record *OperLogEntity) {
	if record == nil {
		return
	}
	if s == nil || s.apiDocSvc == nil {
		return
	}
	text := s.apiDocSvc.ResolveRouteText(ctx, apidoccap.RouteTextInput{
		OperationKey:    record.RouteDocKey,
		Method:          record.RouteMethod,
		Path:            record.RoutePath,
		FallbackTitle:   record.Title,
		FallbackSummary: record.OperSummary,
	})
	record.Title = text.Title
	record.OperSummary = text.Summary
}

// translate resolves one runtime i18n key using the host translation service.
func (s *serviceImpl) translate(ctx context.Context, key string, fallback string) string {
	if s == nil || s.i18nSvc == nil {
		return fallback
	}
	return s.i18nSvc.Translate(ctx, key, fallback)
}

// findLocalizedRouteTitleOperationKeys returns apidoc operation keys whose
// localized module title matches the user's title keyword.
func (s *serviceImpl) findLocalizedRouteTitleOperationKeys(ctx context.Context, title string) []string {
	if s == nil || s.apiDocSvc == nil || strings.TrimSpace(title) == "" {
		return []string{}
	}
	return s.apiDocSvc.FindRouteTitleOperationKeys(ctx, title)
}

// buildStringDictLabelMap builds one localized string-value dictionary label map
// through the host dictionary-domain capability.
func (s *serviceImpl) buildStringDictLabelMap(ctx context.Context, dictType string) map[string]string {
	var (
		fallbacks = defaultOperTypeLabelFallbacks()
		values    = make([]dictcap.Value, 0, len(fallbacks))
		labels    = make(map[string]string, len(fallbacks))
	)
	for _, value := range operlogtype.PublishedValues() {
		values = append(values, dictcap.Value(value))
		labels[value] = s.localizeDictValue(ctx, dictType, value, fallbacks[value])
	}

	result := s.resolveDictLabels(ctx, dictType, values)
	for value, projection := range result {
		if projection != "" {
			labels[string(value)] = projection
		}
	}
	return labels
}

// buildIntDictLabelMap builds one localized integer-value dictionary label map
// through the host dictionary-domain capability.
func (s *serviceImpl) buildIntDictLabelMap(ctx context.Context, dictType string) map[int]string {
	values := make([]dictcap.Value, 0, len(defaultOperStatusLabels))
	labels := make(map[int]string, len(defaultOperStatusLabels))
	for value, fallback := range defaultOperStatusLabels {
		rawValue := strconv.Itoa(value)
		values = append(values, dictcap.Value(rawValue))
		labels[value] = s.localizeDictValue(ctx, dictType, rawValue, fallback)
	}

	result := s.resolveDictLabels(ctx, dictType, values)
	for value, projection := range result {
		parsed, convErr := strconv.Atoi(string(value))
		if convErr != nil {
			continue
		}
		if projection != "" {
			labels[parsed] = projection
		}
	}
	return labels
}

// resolveDictLabels resolves dictionary labels through dictcap and falls back to
// existing runtime i18n keys if the capability is unavailable or misses values.
func (s *serviceImpl) resolveDictLabels(ctx context.Context, dictType string, values []dictcap.Value) map[dictcap.Value]string {
	if s == nil || s.dictSvc == nil || len(values) == 0 {
		return map[dictcap.Value]string{}
	}
	valueSvc := s.dictSvc.Value()
	if valueSvc == nil {
		return map[dictcap.Value]string{}
	}
	output, err := valueSvc.ResolveLabels(ctx, dictcap.ResolveInput{
		Type:         dictcap.Type(dictType),
		Values:       values,
		IncludeLabel: true,
	})
	if err != nil || output == nil || len(output.Items) == 0 {
		return map[dictcap.Value]string{}
	}
	labels := make(map[dictcap.Value]string, len(output.Items))
	for value, projection := range output.Items {
		if projection == nil || strings.TrimSpace(projection.Label) == "" {
			continue
		}
		labels[value] = projection.Label
	}
	return labels
}

// localizeDictValue translates one dictionary label by stable dictionary key.
func (s *serviceImpl) localizeDictValue(ctx context.Context, dictType string, value string, fallback string) string {
	key := strings.Join([]string{dictKeyPrefix, dictType, value, labelKeySuffix}, ".")
	return s.translate(ctx, key, fallback)
}

// defaultOperTypeLabelFallbacks indexes stable operation-type fallback labels by value.
func defaultOperTypeLabelFallbacks() map[string]string {
	labels := make(map[string]string, len(defaultOperTypeLabels))
	for value, label := range defaultOperTypeLabels {
		labels[value.String()] = label
	}
	return labels
}

// defaultOperTypeLabels provides a stable fallback when the dictionary module
// is unavailable during export rendering.
var defaultOperTypeLabels = map[operlogtype.OperType]string{
	operlogtype.OperTypeCreate: "Create",
	operlogtype.OperTypeUpdate: "Update",
	operlogtype.OperTypeDelete: "Delete",
	operlogtype.OperTypeExport: "Export",
	operlogtype.OperTypeImport: "Import",
	operlogtype.OperTypeOther:  "Other",
}

// defaultOperStatusLabels provides a stable fallback when the dictionary module
// is unavailable during export rendering.
var defaultOperStatusLabels = map[int]string{
	OperStatusSuccess: "Success",
	OperStatusFail:    "Failure",
}

// resolveAuditTenantContext resolves tenant audit metadata from bizctx and explicit overrides.
func resolveAuditTenantContext(
	ctx context.Context,
	tenantFilter tenantcap.FilterService,
	tenantID *int,
	actingUserID *int,
	onBehalfOfTenantID *int,
	isImpersonation *bool,
) auditTenantContext {
	if tenantFilter == nil {
		return auditTenantContext{}
	}
	current := tenantFilter.Context(ctx)
	result := auditTenantContext{
		TenantID:           current.TenantID,
		ActingUserID:       current.ActingUserID,
		OnBehalfOfTenantID: current.OnBehalfOfTenantID,
		IsImpersonation:    current.IsImpersonation,
	}
	if tenantID != nil {
		result.TenantID = *tenantID
	}
	if actingUserID != nil {
		result.ActingUserID = *actingUserID
	}
	if onBehalfOfTenantID != nil {
		result.OnBehalfOfTenantID = *onBehalfOfTenantID
	}
	if isImpersonation != nil {
		result.IsImpersonation = *isImpersonation
	}
	return result
}

// applyOperLogFilters wires the shared operation-log query filters onto one model.
func applyOperLogFilters(
	model *gdb.Model,
	title string,
	titleOperationKeys []string,
	operName string,
	operType *operlogtype.OperType,
	status *int,
	beginTime string,
	endTime string,
) *gdb.Model {
	if title != "" {
		if len(titleOperationKeys) > 0 {
			model = model.Where("("+colTitle+" LIKE ? OR "+colRouteDocKey+" IN(?))", "%"+title+"%", titleOperationKeys)
		} else {
			model = model.WhereLike(colTitle, "%"+title+"%")
		}
	}
	if operName != "" {
		model = model.WhereLike(colOperName, "%"+operName+"%")
	}
	if operType != nil {
		model = model.Where(colOperType, operType.String())
	}
	if status != nil {
		model = model.Where(colStatus, *status)
	}
	if beginTime != "" {
		model = model.WhereGTE(colOperTime, beginTime)
	}
	if endTime != "" {
		model = model.WhereLTE(colOperTime, normalizeEndTime(endTime))
	}
	return model
}

// normalizeEndTime expands date-only end values to the end of day.
func normalizeEndTime(value string) string {
	if len(value) == 10 {
		return value + " 23:59:59"
	}
	return value
}
