// This file implements CRUD and attachment storage behavior for the dynamic
// plugin demo-record sample.

package dynamicservice

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/storagecap"
)

// Demo-record constants define the sample table name, paging defaults, and
// attachment limits used by the dynamic plugin sample.
const (
	demoRecordTable                  = "plugin_linapro_demo_dynamic_record"
	demoRecordAttachmentPrefix       = "demo-record-files"
	demoRecordDefaultPageNum         = 1
	demoRecordDefaultPageSize        = 20
	demoRecordMaxPageSize            = 100
	demoRecordMaxAttachmentBytes int = 10 * 1024 * 1024
)

// IsDemoRecordInvalidInput reports whether the error should be exposed as one bad request.
func IsDemoRecordInvalidInput(err error) bool {
	return bizerr.Is(err, CodeDynamicDemoRecordInvalidInput) ||
		bizerr.Is(err, CodeDynamicDemoRecordRequestBodyRequired) ||
		bizerr.Is(err, CodeDynamicDemoRecordTitleRequired) ||
		bizerr.Is(err, CodeDynamicDemoRecordTitleTooLong) ||
		bizerr.Is(err, CodeDynamicDemoRecordContentTooLong) ||
		bizerr.Is(err, CodeDynamicDemoRecordAttachmentNameRequired) ||
		bizerr.Is(err, CodeDynamicDemoRecordAttachmentBase64Invalid) ||
		bizerr.Is(err, CodeDynamicDemoRecordAttachmentContentRequired) ||
		bizerr.Is(err, CodeDynamicDemoRecordAttachmentSizeTooLarge) ||
		bizerr.Is(err, CodeDynamicDemoRecordIDRequired)
}

// NewDemoRecordInvalidInputError creates one invalid-input business error that
// should be translated into a bad-request bridge response.
func NewDemoRecordInvalidInputError(message string) error {
	return bizerr.NewCode(
		CodeDynamicDemoRecordInvalidInput,
		bizerr.P("reason", strings.TrimSpace(message)),
	)
}

// IsDemoRecordNotFound reports whether the error should be exposed as one not-found response.
func IsDemoRecordNotFound(err error) bool {
	return bizerr.Is(err, CodeDynamicDemoRecordNotFound) ||
		bizerr.Is(err, CodeDynamicDemoRecordAttachmentNotFound)
}

// ListDemoRecordsPayload returns one paged demo-record list backed by the plugin-owned SQL table.
func (s *serviceImpl) ListDemoRecordsPayload(input *DemoRecordListInput) (*demoRecordListPayload, error) {
	pageNum, pageSize, keyword := normalizeDemoRecordListInput(input)

	query := s.recordStoreSvc.Table(demoRecordTable).
		Fields("id", "title", "content", "attachmentName", "attachmentPath", "createdAt", "updatedAt").
		OrderDesc("updatedAt").
		OrderDesc("id").
		Page(int32(pageNum), int32(pageSize))
	if keyword != "" {
		query = query.WhereLike("title", keyword)
	}

	records, total, err := query.All()
	if err != nil {
		return nil, err
	}

	items := make([]*demoRecordPayload, 0, len(records))
	for _, record := range records {
		item, mapErr := mapDemoRecordPayload(record)
		if mapErr != nil {
			return nil, mapErr
		}
		items = append(items, item)
	}
	return &demoRecordListPayload{
		List:  items,
		Total: int(total),
	}, nil
}

// GetDemoRecordPayload returns one demo-record detail by ID.
func (s *serviceImpl) GetDemoRecordPayload(recordID string) (*demoRecordPayload, error) {
	recordID, err := validateDemoRecordID(recordID)
	if err != nil {
		return nil, err
	}
	record, err := s.loadDemoRecord(recordID)
	if err != nil {
		return nil, err
	}
	return demoRecordEntityToPayload(record), nil
}

// CreateDemoRecordPayload creates one demo record and stores its optional attachment.
func (s *serviceImpl) CreateDemoRecordPayload(ctx context.Context, input *DemoRecordMutationInput) (payload *demoRecordPayload, err error) {
	normalizedInput, err := normalizeDemoRecordMutationInput(input)
	if err != nil {
		return nil, err
	}
	recordID, err := s.runtimeSvc.UUID()
	if err != nil {
		return nil, err
	}
	attachmentName, attachmentPath, err := s.saveDemoRecordAttachment(ctx, normalizedInput)
	if err != nil {
		return nil, err
	}
	if attachmentPath != "" {
		defer func() {
			if err == nil {
				return
			}
			if cleanupErr := s.deleteDemoRecordAttachment(ctx, attachmentPath); cleanupErr != nil {
				err = bizerr.WrapCode(cleanupErr, CodeDynamicDemoRecordAttachmentRollbackFailed)
			}
		}()
	}
	recordMap, err := buildRecordMap(&demoRecordCreateRecord{
		Id:             recordID,
		Title:          normalizedInput.Title,
		Content:        normalizedInput.Content,
		AttachmentName: attachmentName,
		AttachmentPath: attachmentPath,
	})
	if err != nil {
		return nil, err
	}
	if _, err = s.recordStoreSvc.Table(demoRecordTable).Insert(recordMap); err != nil {
		return nil, err
	}
	record, err := s.loadDemoRecord(recordID)
	if err != nil {
		return nil, err
	}
	payload = demoRecordEntityToPayload(record)
	return payload, nil
}

// UpdateDemoRecordPayload updates one demo record and replaces or removes its optional attachment.
func (s *serviceImpl) UpdateDemoRecordPayload(ctx context.Context, recordID string, input *DemoRecordMutationInput) (payload *demoRecordPayload, err error) {
	validatedRecordID, validateErr := validateDemoRecordID(recordID)
	if validateErr != nil {
		return nil, validateErr
	}
	existingRecord, err := s.loadDemoRecord(validatedRecordID)
	if err != nil {
		return nil, err
	}

	normalizedInput, err := normalizeDemoRecordMutationInput(input)
	if err != nil {
		return nil, err
	}
	newAttachmentName := existingRecord.AttachmentName
	newAttachmentPath := existingRecord.AttachmentPath
	replacedAttachmentPath := ""
	if normalizedInput.RemoveAttachment {
		newAttachmentName = ""
		newAttachmentPath = ""
		replacedAttachmentPath = existingRecord.AttachmentPath
	}

	if strings.TrimSpace(normalizedInput.AttachmentContentBase64) != "" {
		uploadedName, uploadedPath, uploadErr := s.saveDemoRecordAttachment(ctx, normalizedInput)
		if uploadErr != nil {
			return nil, uploadErr
		}
		newAttachmentName = uploadedName
		newAttachmentPath = uploadedPath
		replacedAttachmentPath = existingRecord.AttachmentPath
		defer func() {
			if err == nil || uploadedPath == "" {
				return
			}
			if cleanupErr := s.deleteDemoRecordAttachment(ctx, uploadedPath); cleanupErr != nil {
				err = bizerr.WrapCode(cleanupErr, CodeDynamicDemoRecordAttachmentRollbackFailed)
			}
		}()
	}

	recordMap, err := buildRecordMap(&demoRecordUpdateRecord{
		Title:          normalizedInput.Title,
		Content:        normalizedInput.Content,
		AttachmentName: newAttachmentName,
		AttachmentPath: newAttachmentPath,
	})
	if err != nil {
		return nil, err
	}
	if _, err = s.recordStoreSvc.Table(demoRecordTable).WhereKey(validatedRecordID).Update(recordMap); err != nil {
		return nil, err
	}
	if replacedAttachmentPath != "" && replacedAttachmentPath != newAttachmentPath {
		if err = s.deleteDemoRecordAttachment(ctx, replacedAttachmentPath); err != nil {
			return nil, err
		}
	}

	record, err := s.loadDemoRecord(validatedRecordID)
	if err != nil {
		return nil, err
	}
	payload = demoRecordEntityToPayload(record)
	return payload, nil
}

// DeleteDemoRecordPayload deletes one demo record and its optional attachment.
func (s *serviceImpl) DeleteDemoRecordPayload(ctx context.Context, recordID string) (*demoRecordDeletePayload, error) {
	recordID, err := validateDemoRecordID(recordID)
	if err != nil {
		return nil, err
	}
	record, err := s.loadDemoRecord(recordID)
	if err != nil {
		return nil, err
	}
	if _, err = s.recordStoreSvc.Table(demoRecordTable).WhereKey(recordID).Delete(); err != nil {
		return nil, err
	}
	if err = s.deleteDemoRecordAttachment(ctx, record.AttachmentPath); err != nil {
		return nil, err
	}
	return &demoRecordDeletePayload{
		Id:      recordID,
		Deleted: true,
	}, nil
}

// BuildDemoRecordAttachmentDownload returns one attachment download descriptor.
func (s *serviceImpl) BuildDemoRecordAttachmentDownload(ctx context.Context, recordID string) (*demoRecordAttachmentDownloadPayload, error) {
	recordID, err := validateDemoRecordID(recordID)
	if err != nil {
		return nil, err
	}
	record, err := s.loadDemoRecord(recordID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(record.AttachmentPath) == "" {
		return nil, bizerr.NewCode(CodeDynamicDemoRecordAttachmentNotFound)
	}

	output, err := s.storageSvc.Get(ctx, storagecap.GetInput{Path: record.AttachmentPath})
	if err != nil {
		return nil, err
	}
	if output == nil || !output.Found {
		return nil, bizerr.NewCode(CodeDynamicDemoRecordAttachmentNotFound)
	}
	body, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, err
	}
	if output.Body != nil {
		if closeErr := output.Body.Close(); closeErr != nil {
			return nil, closeErr
		}
	}

	contentType := "application/octet-stream"
	if output.Object != nil && strings.TrimSpace(output.Object.ContentType) != "" {
		contentType = strings.TrimSpace(output.Object.ContentType)
	}
	return &demoRecordAttachmentDownloadPayload{
		OriginalName: record.AttachmentName,
		ContentType:  contentType,
		Body:         body,
	}, nil
}

// loadDemoRecord loads one sample record entity by its logical ID.
func (s *serviceImpl) loadDemoRecord(recordID string) (*demoRecordEntity, error) {
	recordMap, found, err := s.recordStoreSvc.Table(demoRecordTable).
		Fields("id", "title", "content", "attachmentName", "attachmentPath", "createdAt", "updatedAt").
		WhereKey(recordID).
		One()
	if err != nil {
		return nil, err
	}
	if !found || recordMap == nil {
		return nil, bizerr.NewCode(CodeDynamicDemoRecordNotFound)
	}
	return parseDemoRecordEntity(recordMap)
}

// mapDemoRecordPayload converts one raw structured-data row into the response
// payload shape.
func mapDemoRecordPayload(record map[string]any) (*demoRecordPayload, error) {
	entity, err := parseDemoRecordEntity(record)
	if err != nil {
		return nil, err
	}
	return demoRecordEntityToPayload(entity), nil
}

// demoRecordEntityToPayload converts one internal record entity into the JSON
// payload returned by the bridge controller.
func demoRecordEntityToPayload(record *demoRecordEntity) *demoRecordPayload {
	if record == nil {
		return &demoRecordPayload{}
	}
	return &demoRecordPayload{
		Id:             record.Id,
		Title:          record.Title,
		Content:        record.Content,
		AttachmentName: record.AttachmentName,
		HasAttachment:  strings.TrimSpace(record.AttachmentPath) != "",
		CreatedAt:      demoRecordMilliFromString(record.CreatedAt),
		UpdatedAt:      demoRecordMilliFromString(record.UpdatedAt),
	}
}

// demoRecordMilliFromString converts table timestamp strings into Unix
// milliseconds without using GoFrame time parsing inside the WASM guest.
func demoRecordMilliFromString(value string) *int64 {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	parsed, ok := parseDemoRecordTimestamp(trimmed)
	if !ok || parsed.IsZero() {
		return nil
	}
	millis := parsed.UnixMilli()
	return &millis
}

// parseDemoRecordTimestamp parses the stable timestamp prefixes returned by
// PostgreSQL through GoFrame without invoking generic time parsers in WASM.
func parseDemoRecordTimestamp(value string) (time.Time, bool) {
	if len(value) < len(demoRecordTimestampLayoutPrefix) {
		return time.Time{}, false
	}
	if value[4] != '-' || value[7] != '-' || (value[10] != ' ' && value[10] != 'T') ||
		value[13] != ':' || value[16] != ':' {
		return time.Time{}, false
	}

	year, ok := parseDemoRecordTimestampPart(value[0:4])
	if !ok {
		return time.Time{}, false
	}
	month, ok := parseDemoRecordTimestampPart(value[5:7])
	if !ok {
		return time.Time{}, false
	}
	day, ok := parseDemoRecordTimestampPart(value[8:10])
	if !ok {
		return time.Time{}, false
	}
	hour, ok := parseDemoRecordTimestampPart(value[11:13])
	if !ok {
		return time.Time{}, false
	}
	minute, ok := parseDemoRecordTimestampPart(value[14:16])
	if !ok {
		return time.Time{}, false
	}
	second, ok := parseDemoRecordTimestampPart(value[17:19])
	if !ok {
		return time.Time{}, false
	}

	nanosecond, offsetSeconds, offsetSet, ok := parseDemoRecordTimestampSuffix(value[19:])
	if !ok {
		return time.Time{}, false
	}
	location := time.Local
	if offsetSet {
		location = time.FixedZone("", offsetSeconds)
	}
	return time.Date(year, time.Month(month), day, hour, minute, second, nanosecond, location), true
}

// parseDemoRecordTimestampPart parses one zero-padded numeric timestamp part.
func parseDemoRecordTimestampPart(value string) (int, bool) {
	if value == "" {
		return 0, false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return 0, false
		}
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

// parseDemoRecordTimestampSuffix parses optional fractional seconds and
// timezone suffixes for database timestamp strings.
func parseDemoRecordTimestampSuffix(value string) (int, int, bool, bool) {
	rest := strings.TrimSpace(value)
	nanosecond := 0
	if strings.HasPrefix(rest, ".") {
		digitsEnd := 1
		for digitsEnd < len(rest) {
			r := rest[digitsEnd]
			if r < '0' || r > '9' {
				break
			}
			digitsEnd++
		}
		fraction := rest[1:digitsEnd]
		if fraction == "" {
			return 0, 0, false, false
		}
		if len(fraction) > 9 {
			fraction = fraction[:9]
		}
		for len(fraction) < 9 {
			fraction += "0"
		}
		parsed, ok := parseDemoRecordTimestampPart(fraction)
		if !ok {
			return 0, 0, false, false
		}
		nanosecond = parsed
		rest = strings.TrimSpace(rest[digitsEnd:])
	}
	if rest == "" {
		return nanosecond, 0, false, true
	}
	if rest == "Z" || rest == "z" {
		return nanosecond, 0, true, true
	}
	offsetSeconds, ok := parseDemoRecordTimezoneOffset(rest)
	if !ok {
		return 0, 0, false, false
	}
	return nanosecond, offsetSeconds, true, true
}

// parseDemoRecordTimezoneOffset parses +08, +0800, and +08:00 suffixes.
func parseDemoRecordTimezoneOffset(value string) (int, bool) {
	if len(value) != 3 && len(value) != 5 && len(value) != 6 {
		return 0, false
	}
	sign := 1
	switch value[0] {
	case '+':
	case '-':
		sign = -1
	default:
		return 0, false
	}
	hourText := value[1:3]
	minuteText := "00"
	if len(value) == 5 {
		minuteText = value[3:5]
	}
	if len(value) == 6 {
		if value[3] != ':' {
			return 0, false
		}
		minuteText = value[4:6]
	}
	hour, ok := parseDemoRecordTimestampPart(hourText)
	if !ok {
		return 0, false
	}
	minute, ok := parseDemoRecordTimestampPart(minuteText)
	if !ok {
		return 0, false
	}
	if hour > 23 || minute > 59 {
		return 0, false
	}
	return sign * ((hour * 60 * 60) + (minute * 60)), true
}

// normalizeDemoRecordListInput applies paging defaults and trims the keyword
// filter for list requests.
func normalizeDemoRecordListInput(input *DemoRecordListInput) (int, int, string) {
	pageNum := demoRecordDefaultPageNum
	pageSize := demoRecordDefaultPageSize
	keyword := ""
	if input == nil {
		return pageNum, pageSize, keyword
	}
	if input.PageNum > 0 {
		pageNum = input.PageNum
	}
	if input.PageSize > 0 {
		pageSize = input.PageSize
	}
	if pageSize > demoRecordMaxPageSize {
		pageSize = demoRecordMaxPageSize
	}
	keyword = strings.TrimSpace(input.Keyword)
	return pageNum, pageSize, keyword
}

// normalizeDemoRecordMutationInput validates and trims create/update request
// data for sample records.
func normalizeDemoRecordMutationInput(input *DemoRecordMutationInput) (*DemoRecordMutationInput, error) {
	if input == nil {
		return nil, bizerr.NewCode(CodeDynamicDemoRecordRequestBodyRequired)
	}

	normalizedInput := &DemoRecordMutationInput{
		Title:                   strings.TrimSpace(input.Title),
		Content:                 strings.TrimSpace(input.Content),
		AttachmentName:          strings.TrimSpace(input.AttachmentName),
		AttachmentContentBase64: strings.TrimSpace(input.AttachmentContentBase64),
		AttachmentContentType:   strings.TrimSpace(input.AttachmentContentType),
		RemoveAttachment:        input.RemoveAttachment,
	}

	if normalizedInput.Title == "" {
		return nil, bizerr.NewCode(CodeDynamicDemoRecordTitleRequired)
	}
	if len([]rune(normalizedInput.Title)) > 128 {
		return nil, bizerr.NewCode(CodeDynamicDemoRecordTitleTooLong, bizerr.P("maxChars", 128))
	}
	if len([]rune(normalizedInput.Content)) > 1000 {
		return nil, bizerr.NewCode(CodeDynamicDemoRecordContentTooLong, bizerr.P("maxChars", 1000))
	}
	if normalizedInput.AttachmentContentBase64 != "" && normalizedInput.AttachmentName == "" {
		return nil, bizerr.NewCode(CodeDynamicDemoRecordAttachmentNameRequired)
	}
	return normalizedInput, nil
}

// saveDemoRecordAttachment decodes and stores one optional Base64 attachment
// into the governed plugin storage area.
func (s *serviceImpl) saveDemoRecordAttachment(ctx context.Context, input *DemoRecordMutationInput) (string, string, error) {
	if input == nil || strings.TrimSpace(input.AttachmentContentBase64) == "" {
		return "", "", nil
	}

	body, err := base64.StdEncoding.DecodeString(input.AttachmentContentBase64)
	if err != nil {
		return "", "", bizerr.NewCode(CodeDynamicDemoRecordAttachmentBase64Invalid)
	}
	if len(body) == 0 {
		return "", "", bizerr.NewCode(CodeDynamicDemoRecordAttachmentContentRequired)
	}
	if len(body) > demoRecordMaxAttachmentBytes {
		return "", "", bizerr.NewCode(
			CodeDynamicDemoRecordAttachmentSizeTooLarge,
			bizerr.P("maxSizeMB", demoRecordMaxAttachmentBytes/1024/1024),
		)
	}

	attachmentName := sanitizeDemoRecordAttachmentName(input.AttachmentName)
	attachmentID, err := s.runtimeSvc.UUID()
	if err != nil {
		return "", "", err
	}

	objectPath := fmt.Sprintf(
		"%s/%s_%s",
		demoRecordAttachmentPrefix,
		attachmentID,
		attachmentName,
	)
	contentType := strings.TrimSpace(input.AttachmentContentType)
	if contentType == "" {
		contentType = http.DetectContentType(body)
	}
	if _, err = s.storageSvc.Put(ctx, storagecap.PutInput{
		Path:        objectPath,
		Body:        bytes.NewReader(body),
		Size:        int64(len(body)),
		ContentType: contentType,
		Overwrite:   true,
	}); err != nil {
		return "", "", err
	}
	return attachmentName, objectPath, nil
}

// deleteDemoRecordAttachment removes one governed attachment object when its
// logical path is present.
func (s *serviceImpl) deleteDemoRecordAttachment(ctx context.Context, objectPath string) error {
	if strings.TrimSpace(objectPath) == "" {
		return nil
	}
	return s.storageSvc.Delete(ctx, storagecap.DeleteInput{Path: objectPath})
}

// validateDemoRecordID validates the logical record identifier used by sample
// CRUD operations.
func validateDemoRecordID(value string) (string, error) {
	normalizedValue := strings.TrimSpace(value)
	if normalizedValue == "" {
		return "", bizerr.NewCode(CodeDynamicDemoRecordIDRequired)
	}
	return normalizedValue, nil
}

// sanitizeDemoRecordAttachmentName strips unsafe path characters from uploaded
// attachment names.
func sanitizeDemoRecordAttachmentName(filename string) string {
	sanitizedName := filepath.Base(strings.ReplaceAll(strings.TrimSpace(filename), "\x00", ""))
	if sanitizedName == "." || sanitizedName == "" {
		return "attachment.bin"
	}

	disallowed := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, item := range disallowed {
		sanitizedName = strings.ReplaceAll(sanitizedName, item, "_")
	}
	return sanitizedName
}
