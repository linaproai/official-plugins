// loginlog_impl.go implements login-log persistence, tenant-filtered queries,
// cleanup, and Excel export for the linapro-monitor-loginlog plugin. It resolves
// dictionary and runtime i18n labels while keeping plugin table access scoped
// through the injected tenant filter service.

package loginlog

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
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/dictcap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/pluginhost"
	"lina-plugin-linapro-monitor-loginlog/backend/internal/dao"
	"lina-plugin-linapro-monitor-loginlog/backend/internal/model/do"
)

// Create inserts one login-log record.
func (s *serviceImpl) Create(ctx context.Context, in CreateInput) error {
	auditContext := resolveAuditTenantContext(
		ctx,
		s.tenantFilter,
		in.TenantID,
		in.ActingUserID,
		in.OnBehalfOfTenantID,
		in.IsImpersonation,
	)

	_, err := dao.Loginlog.Ctx(ctx).Data(do.Loginlog{
		TenantId:           auditContext.TenantID,
		ActingUserId:       auditContext.ActingUserID,
		OnBehalfOfTenantId: auditContext.OnBehalfOfTenantID,
		IsImpersonation:    auditContext.IsImpersonation,
		UserName:           in.UserName,
		Status:             in.Status,
		Ip:                 in.Ip,
		Browser:            in.Browser,
		Os:                 in.Os,
		Msg:                in.Msg,
		LoginTime:          timePtr(time.Now()),
	}).Insert()
	return err
}

// timePtr returns a pointer to value for generated DO time fields that preserve
// database NULL semantics with *time.Time.
func timePtr(value time.Time) *time.Time {
	return &value
}

// List queries the paginated login-log list.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	model := s.tenantFilter.Apply(ctx, dao.Loginlog.Ctx(ctx), "")
	model = applyLoginLogFilters(model, in.UserName, in.Ip, in.Status, in.BeginTime, in.EndTime)

	total, err := model.Count()
	if err != nil {
		return nil, err
	}

	allowedSortFields := map[string]string{
		"id":         colID,
		"loginTime":  colLoginTime,
		"login_time": colLoginTime,
	}
	orderBy := colLoginTime
	if field, ok := allowedSortFields[in.OrderBy]; ok {
		orderBy = field
	}
	direction := gdbutil.NormalizeOrderDirectionOrDefault(in.OrderDirection, gdbutil.OrderDirectionDESC)

	list := make([]*LoginLogEntity, 0)
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

// GetById retrieves one login-log record by primary key.
func (s *serviceImpl) GetById(ctx context.Context, id int) (*LoginLogEntity, error) {
	var record *LoginLogEntity
	err := s.tenantFilter.Apply(ctx, dao.Loginlog.Ctx(ctx), "").Where(colID, id).Scan(&record)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, bizerr.NewCode(CodeLoginLogNotFound)
	}
	s.localizeRecord(ctx, record)
	return record, nil
}

// Clean hard-deletes login logs within one optional time range.
func (s *serviceImpl) Clean(ctx context.Context, in CleanInput) (int, error) {
	model := s.tenantFilter.Apply(ctx, dao.Loginlog.Ctx(ctx), "")
	hasFilter := false
	if in.BeginTime != "" {
		model = model.WhereGTE(colLoginTime, in.BeginTime)
		hasFilter = true
	}
	if in.EndTime != "" {
		model = model.WhereLTE(colLoginTime, normalizeEndTime(in.EndTime))
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

// CleanupExpired hard-deletes login logs older than the global retention boundary.
func (s *serviceImpl) CleanupExpired(ctx context.Context, retentionDays int) (int, error) {
	if retentionDays <= 0 {
		return 0, nil
	}
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	result, err := dao.Loginlog.Ctx(ctx).
		WhereLT(colLoginTime, cutoff).
		Delete()
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(affected), nil
}

// DeleteByIds hard-deletes login logs by ID list.
func (s *serviceImpl) DeleteByIds(ctx context.Context, ids []int) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result, err := s.tenantFilter.Apply(ctx, dao.Loginlog.Ctx(ctx), "").WhereIn(colID, ids).Delete()
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(affected), nil
}

// Export generates an Excel workbook for login logs.
func (s *serviceImpl) Export(ctx context.Context, in ExportInput) (data []byte, err error) {
	model := s.tenantFilter.Apply(ctx, dao.Loginlog.Ctx(ctx), "")
	if len(in.Ids) > 0 {
		model = model.WhereIn(colID, in.Ids)
	} else {
		model = applyLoginLogFilters(model, in.UserName, in.Ip, in.Status, in.BeginTime, in.EndTime)
	}
	model = model.Limit(MaxExportRows)

	allowedSortFields := map[string]string{
		"id":         colID,
		"loginTime":  colLoginTime,
		"login_time": colLoginTime,
	}
	orderBy := colLoginTime
	if field, ok := allowedSortFields[in.OrderBy]; ok {
		orderBy = field
	}
	direction := gdbutil.NormalizeOrderDirectionOrDefault(in.OrderDirection, gdbutil.OrderDirectionDESC)

	list := make([]*LoginLogEntity, 0)
	err = gdbutil.ApplyModelOrder(model, orderBy, direction).Scan(&list)
	if err != nil {
		return nil, err
	}

	file := excelize.NewFile()
	defer excelutil.CloseFile(ctx, file, &err)
	sheet := "Sheet1"
	headers := s.exportHeaders(ctx)
	for index, header := range headers {
		if setErr := excelutil.SetCellValue(file, sheet, index+1, 1, header); setErr != nil {
			return nil, setErr
		}
	}

	statusMap := s.buildIntDictLabelMap(ctx, DictTypeLoginStatus)
	for index, log := range list {
		row := index + 2
		if setErr := excelutil.SetCellValue(file, sheet, 1, row, log.UserName); setErr != nil {
			return nil, setErr
		}
		statusText := s.exportStatusText(ctx, log.Status, statusMap)
		if setErr := excelutil.SetCellValue(file, sheet, 2, row, statusText); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValue(file, sheet, 3, row, log.Ip); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValue(file, sheet, 4, row, log.Browser); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValue(file, sheet, 5, row, log.Os); setErr != nil {
			return nil, setErr
		}
		if setErr := excelutil.SetCellValue(file, sheet, 6, row, s.translateLoginLogMessage(ctx, log.Msg)); setErr != nil {
			return nil, setErr
		}
		if log.LoginTime != nil {
			if setErr := excelutil.SetCellValue(file, sheet, 7, row, log.LoginTime.String()); setErr != nil {
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

// exportHeaders returns localized Excel headers for login-log export.
func (s *serviceImpl) exportHeaders(ctx context.Context) []string {
	return []string{
		s.translate(ctx, "plugin.linapro-monitor-loginlog.fields.userName", "User Account"),
		s.translate(ctx, "plugin.linapro-monitor-loginlog.fields.status", "Login Status"),
		s.translate(ctx, "plugin.linapro-monitor-loginlog.fields.ipAddress", "IP Address"),
		s.translate(ctx, "plugin.linapro-monitor-loginlog.fields.browser", "Browser"),
		s.translate(ctx, "plugin.linapro-monitor-loginlog.fields.os", "Operating System"),
		s.translate(ctx, "plugin.linapro-monitor-loginlog.fields.message", "Message"),
		s.translate(ctx, "plugin.linapro-monitor-loginlog.fields.loginTime", "Login Time"),
	}
}

// exportStatusText returns the localized export label for one login status.
func (s *serviceImpl) exportStatusText(ctx context.Context, status int, statusMap map[int]string) string {
	statusText, ok := statusMap[status]
	if !ok {
		statusText = defaultLoginStatusLabels[status]
	}
	return s.translateDictLabel(ctx, DictTypeLoginStatus, strconv.Itoa(status), statusText)
}

// localizeRecords translates backend-owned display fields for login-log rows.
func (s *serviceImpl) localizeRecords(ctx context.Context, records []*LoginLogEntity) {
	for _, record := range records {
		s.localizeRecord(ctx, record)
	}
}

// localizeRecord translates backend-owned display fields for one login-log row.
func (s *serviceImpl) localizeRecord(ctx context.Context, record *LoginLogEntity) {
	if record == nil {
		return
	}
	record.Msg = s.translateLoginLogMessage(ctx, record.Msg)
}

// translateLoginLogMessage resolves stable auth lifecycle reason codes.
func (s *serviceImpl) translateLoginLogMessage(ctx context.Context, message string) string {
	key := loginLogReasonI18nKey(strings.TrimSpace(message))
	if key == "" {
		return message
	}
	return s.translate(ctx, key, message)
}

// translateDictLabel translates one dictionary label through runtime i18n keys.
func (s *serviceImpl) translateDictLabel(ctx context.Context, dictType string, value string, fallback string) string {
	key := strings.Join([]string{dictKeyPrefix, dictType, value, labelKeySuffix}, ".")
	return s.translate(ctx, key, fallback)
}

// translate resolves one runtime i18n key through the host i18n service.
func (s *serviceImpl) translate(ctx context.Context, key string, fallback string) string {
	if s == nil || s.i18nSvc == nil || strings.TrimSpace(key) == "" {
		return fallback
	}
	return s.i18nSvc.Translate(ctx, key, fallback)
}

var defaultLoginStatusLabels = map[int]string{
	LoginStatusSuccess: "Success",
	LoginStatusFail:    "Failed",
}

// resolveAuditTenantContext resolves tenant audit metadata from bizctx and explicit overrides.
func resolveAuditTenantContext(
	ctx context.Context,
	tenantFilter tenantcap.PluginTableFilterService,
	tenantID *int,
	actingUserID *int,
	onBehalfOfTenantID *int,
	isImpersonation *bool,
) auditTenantContext {
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

// loginLogReasonI18nKey maps published auth reason codes to plugin-owned i18n keys.
func loginLogReasonI18nKey(reason string) string {
	switch reason {
	case pluginhost.AuthHookReasonLoginSuccessful:
		return loginLogMessagePrefix + ".loginSuccessful"
	case pluginhost.AuthHookReasonLoginFailed:
		return loginLogMessagePrefix + ".loginFailed"
	case pluginhost.AuthHookReasonLogoutSuccessful:
		return loginLogMessagePrefix + ".logoutSuccessful"
	case pluginhost.AuthHookReasonInvalidCredentials:
		return loginLogMessagePrefix + ".invalidCredentials"
	case pluginhost.AuthHookReasonUserDisabled:
		return loginLogMessagePrefix + ".userDisabled"
	case pluginhost.AuthHookReasonIPBlacklisted:
		return loginLogMessagePrefix + ".ipBlacklisted"
	}
	return ""
}

// applyLoginLogFilters wires the shared login-log query filters onto one model.
func applyLoginLogFilters(model *gdb.Model, userName string, ip string, status *int, beginTime string, endTime string) *gdb.Model {
	if userName != "" {
		model = model.WhereLike(colUserName, "%"+userName+"%")
	}
	if ip != "" {
		model = model.WhereLike(colIP, "%"+ip+"%")
	}
	if status != nil {
		model = model.Where(colStatus, *status)
	}
	if beginTime != "" {
		model = model.WhereGTE(colLoginTime, beginTime)
	}
	if endTime != "" {
		model = model.WhereLTE(colLoginTime, normalizeEndTime(endTime))
	}
	return model
}

// buildIntDictLabelMap builds one integer-value dictionary label map through
// the host dictionary-domain capability.
func (s *serviceImpl) buildIntDictLabelMap(ctx context.Context, dictType string) map[int]string {
	values := make([]dictcap.Value, 0, len(defaultLoginStatusLabels))
	labels := make(map[int]string, len(defaultLoginStatusLabels))
	for value, fallback := range defaultLoginStatusLabels {
		rawValue := strconv.Itoa(value)
		values = append(values, dictcap.Value(rawValue))
		labels[value] = s.translateDictLabel(ctx, dictType, rawValue, fallback)
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

// resolveDictLabels resolves dictionary labels through dictcap and leaves
// missing values to the existing runtime i18n fallback path.
func (s *serviceImpl) resolveDictLabels(ctx context.Context, dictType string, values []dictcap.Value) map[dictcap.Value]string {
	if s == nil || s.dictSvc == nil || len(values) == 0 {
		return map[dictcap.Value]string{}
	}
	output, err := s.dictSvc.ResolveLabels(ctx, s.dictionaryCapabilityContext(ctx, dictType), dictcap.ResolveInput{
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

// dictionaryCapabilityContext builds the audited context required by dictcap.
func (s *serviceImpl) dictionaryCapabilityContext(ctx context.Context, dictType string) capmodel.CapabilityContext {
	current := tenantcap.TenantFilterContext{}
	if s != nil && s.tenantFilter != nil {
		current = s.tenantFilter.Context(ctx)
	}
	actorID := current.UserID
	if current.ActingUserID > 0 {
		actorID = current.ActingUserID
	}
	actor := capmodel.CapabilityActor{
		Type:   capmodel.ActorTypeUser,
		UserID: int64(actorID),
	}
	if actorID == 0 {
		actor = capmodel.CapabilityActor{
			Type:         capmodel.ActorTypeSystem,
			Name:         pluginID,
			SystemReason: "login-log dictionary label projection",
		}
	}
	return capmodel.CapabilityContext{
		PluginID:   pluginID,
		Actor:      actor,
		TenantID:   capmodel.DomainID(strconv.Itoa(current.TenantID)),
		Source:     capmodel.CapabilitySourceHTTP,
		SystemCall: actor.Type == capmodel.ActorTypeSystem,
		Resource:   dictType,
		AuditReason: strings.Join([]string{
			pluginID,
			"resolve dictionary labels",
			dictType,
		}, ":"),
		RequestedAt: time.Now(),
	}
}

// normalizeEndTime expands date-only end values to the end of day.
func normalizeEndTime(value string) string {
	if len(value) == 10 {
		return value + " 23:59:59"
	}
	return value
}
