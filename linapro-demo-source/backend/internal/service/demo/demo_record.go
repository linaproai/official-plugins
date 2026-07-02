// demo_record.go implements CRUD, paging, and attachment download behavior for
// the linapro-demo-source record sample.

package demo

import (
	"context"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"lina-plugin-linapro-demo-source/backend/internal/dao"
	"lina-plugin-linapro-demo-source/backend/internal/model/do"
	entitymodel "lina-plugin-linapro-demo-source/backend/internal/model/entity"
)

// Demo-record constants define the table schema fields and paging defaults
// used by the source-plugin sample service.
const (
	defaultPageNum  = 1
	defaultPageSize = 10
	maxPageSize     = 100
)

// ListRecordsInput defines the demo record list query.
type ListRecordsInput struct {
	// Keyword is the optional fuzzy-match keyword applied to title.
	Keyword string
	// PageNum is the requested page number.
	PageNum int
	// PageSize is the requested page size.
	PageSize int
}

// ListRecordsOutput defines the demo record list result.
type ListRecordsOutput struct {
	// List contains the current page of records.
	List []*RecordListItemOutput
	// Total is the total matched row count.
	Total int
}

// RecordListItemOutput defines one demo record row.
type RecordListItemOutput struct {
	// Id is the record ID.
	Id int64
	// Title is the record title.
	Title string
	// Content is the record content summary.
	Content string
	// AttachmentName is the original attachment filename.
	AttachmentName string
	// HasAttachment reports whether the record owns one attachment.
	HasAttachment bool
	// CreatedAt is the creation time as a Unix timestamp in milliseconds.
	CreatedAt *int64
	// UpdatedAt is the update time as a Unix timestamp in milliseconds.
	UpdatedAt *int64
}

// RecordDetailOutput defines one demo record detail result.
type RecordDetailOutput struct {
	// Id is the record ID.
	Id int64
	// Title is the record title.
	Title string
	// Content is the record content body.
	Content string
	// AttachmentName is the original attachment filename.
	AttachmentName string
	// HasAttachment reports whether the record owns one attachment.
	HasAttachment bool
}

// CreateRecordInput defines the create-record input.
type CreateRecordInput struct {
	// Title is the required record title.
	Title string
	// Content is the optional record content.
	Content string
	// File is the optional uploaded attachment.
	File *ghttp.UploadFile
}

// UpdateRecordInput defines the update-record input.
type UpdateRecordInput struct {
	// Id is the record ID.
	Id int64
	// Title is the required record title.
	Title string
	// Content is the optional record content.
	Content string
	// File is the optional new uploaded attachment.
	File *ghttp.UploadFile
	// RemoveAttachment reports whether the current attachment should be removed.
	RemoveAttachment bool
}

// RecordMutationOutput defines the record create/update result.
type RecordMutationOutput struct {
	// Id is the affected record ID.
	Id int64
}

// AttachmentDownloadOutput defines one attachment download descriptor.
type AttachmentDownloadOutput struct {
	// OriginalName is the original attachment filename.
	OriginalName string
	// ContentType is the detected content type.
	ContentType string
	// Body carries attachment content. Callers must close it.
	Body io.ReadCloser
	// Size is the attachment object size in bytes when known.
	Size int64
}

// demoRecordEntity reuses the plugin-local generated record entity.
type demoRecordEntity = entitymodel.Record

// ListRecords returns the paged demo records rendered by the source-plugin CRUD page.
func (s *serviceImpl) ListRecords(ctx context.Context, in *ListRecordsInput) (out *ListRecordsOutput, err error) {
	if err = ensureDemoRecordTableReady(ctx); err != nil {
		return nil, err
	}

	pageNum, pageSize := normalizeListPagination(in)
	var (
		cols    = dao.Record.Columns()
		model   = tenantspi.ApplyPluginTableFilter(ctx, s.pluginTableFilter(), dao.Record.Ctx(ctx), "")
		keyword = strings.TrimSpace(in.Keyword)
	)
	if keyword != "" {
		model = model.WhereLike(cols.Title, "%"+keyword+"%")
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeDemoRecordCountQueryFailed)
	}

	items := make([]*demoRecordEntity, 0)
	err = model.
		OrderDesc(cols.UpdatedAt).
		OrderDesc(cols.Id).
		Page(pageNum, pageSize).
		Scan(&items)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeDemoRecordListQueryFailed)
	}

	list := make([]*RecordListItemOutput, 0, len(items))
	for _, item := range items {
		list = append(list, buildRecordListItemOutput(item))
	}
	return &ListRecordsOutput{List: list, Total: total}, nil
}

// GetRecord returns one demo record detail for edit forms.
func (s *serviceImpl) GetRecord(ctx context.Context, id int64) (out *RecordDetailOutput, err error) {
	record, err := s.getRecordEntity(ctx, id)
	if err != nil {
		return nil, err
	}
	return &RecordDetailOutput{
		Id:             record.Id,
		Title:          record.Title,
		Content:        record.Content,
		AttachmentName: record.AttachmentName,
		HasAttachment:  record.AttachmentPath != "",
	}, nil
}

// CreateRecord creates one demo record and stores its optional attachment file.
func (s *serviceImpl) CreateRecord(ctx context.Context, in *CreateRecordInput) (out *RecordMutationOutput, err error) {
	if err = ensureDemoRecordTableReady(ctx); err != nil {
		return nil, err
	}
	if err = validateRecordTitle(in.Title); err != nil {
		return nil, err
	}

	attachmentName, attachmentPath, err := s.saveDemoAttachmentFile(ctx, in.File)
	if err != nil {
		return nil, err
	}
	if attachmentPath != "" {
		defer func() {
			if err != nil {
				s.cleanupDemoAttachmentAfterMutationFailure(ctx, attachmentPath)
			}
		}()
	}

	tenantID := s.tenantFilterContext(ctx).TenantID
	recordID, err := dao.Record.Ctx(ctx).Data(do.Record{
		TenantId:       tenantID,
		Title:          strings.TrimSpace(in.Title),
		Content:        strings.TrimSpace(in.Content),
		AttachmentName: stringPointer(attachmentName),
		AttachmentPath: stringPointer(attachmentPath),
	}).InsertAndGetId()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeDemoRecordCreateFailed)
	}
	return &RecordMutationOutput{Id: recordID}, nil
}

// UpdateRecord updates one demo record and replaces or removes its optional attachment.
func (s *serviceImpl) UpdateRecord(ctx context.Context, in *UpdateRecordInput) (out *RecordMutationOutput, err error) {
	if err = ensureDemoRecordTableReady(ctx); err != nil {
		return nil, err
	}
	if err = validateRecordTitle(in.Title); err != nil {
		return nil, err
	}

	record, err := s.getRecordEntity(ctx, in.Id)
	if err != nil {
		return nil, err
	}

	updateData := do.Record{
		Title:          strings.TrimSpace(in.Title),
		Content:        strings.TrimSpace(in.Content),
		AttachmentName: stringPointer(record.AttachmentName),
		AttachmentPath: stringPointer(record.AttachmentPath),
	}
	oldAttachmentPath := strings.TrimSpace(record.AttachmentPath)

	if in.RemoveAttachment {
		updateData.AttachmentName = stringPointer("")
		updateData.AttachmentPath = stringPointer("")
	}

	newAttachmentPath := ""
	if in.File != nil {
		var newAttachmentName string
		newAttachmentName, newAttachmentPath, err = s.saveDemoAttachmentFile(ctx, in.File)
		if err != nil {
			return nil, err
		}
		updateData.AttachmentName = stringPointer(newAttachmentName)
		updateData.AttachmentPath = stringPointer(newAttachmentPath)
		defer func() {
			if err != nil && newAttachmentPath != "" {
				s.cleanupDemoAttachmentAfterMutationFailure(ctx, newAttachmentPath)
			}
		}()
	}

	tenantID := s.tenantFilterContext(ctx).TenantID
	_, err = dao.Record.Ctx(ctx).
		Where(tenantspi.TenantFilterColumn, tenantID).
		Where(do.Record{Id: in.Id}).
		Data(updateData).
		Update()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeDemoRecordUpdateFailed)
	}

	if (in.RemoveAttachment || newAttachmentPath != "") && oldAttachmentPath != "" {
		if err = s.deleteDemoAttachmentFile(ctx, oldAttachmentPath); err != nil {
			return nil, err
		}
	}
	return &RecordMutationOutput{Id: in.Id}, nil
}

// DeleteRecord deletes one demo record and cleans its attachment file.
func (s *serviceImpl) DeleteRecord(ctx context.Context, id int64) error {
	record, err := s.getRecordEntity(ctx, id)
	if err != nil {
		return err
	}

	tenantID := s.tenantFilterContext(ctx).TenantID
	_, err = dao.Record.Ctx(ctx).
		Where(tenantspi.TenantFilterColumn, tenantID).
		Where(do.Record{Id: id}).
		Delete()
	if err != nil {
		return bizerr.WrapCode(err, CodeDemoRecordDeleteFailed)
	}
	if record.AttachmentPath != "" {
		if err = s.deleteDemoAttachmentFile(ctx, record.AttachmentPath); err != nil {
			return err
		}
	}
	return nil
}

// BuildAttachmentDownload returns one attachment download descriptor for the given record.
func (s *serviceImpl) BuildAttachmentDownload(
	ctx context.Context,
	id int64,
) (out *AttachmentDownloadOutput, err error) {
	record, err := s.getRecordEntity(ctx, id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(record.AttachmentPath) == "" {
		return nil, bizerr.NewCode(CodeDemoRecordAttachmentRequired)
	}

	if s.storageSvc == nil {
		return nil, bizerr.NewCode(CodeDemoAttachmentStorageUnavailable)
	}
	objectOutput, err := s.storageSvc.Get(ctx, storagecap.GetInput{Path: record.AttachmentPath})
	if err != nil {
		return nil, err
	}
	if objectOutput == nil || !objectOutput.Found || objectOutput.Body == nil {
		return nil, bizerr.NewCode(CodeDemoRecordAttachmentFileNotFound)
	}

	contentType := ""
	size := int64(0)
	if objectOutput.Object != nil {
		contentType = strings.TrimSpace(objectOutput.Object.ContentType)
		size = objectOutput.Object.Size
	}
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(record.AttachmentName))
	}
	if contentType == "" {
		contentType = http.DetectContentType(nil)
	}
	return &AttachmentDownloadOutput{
		OriginalName: record.AttachmentName,
		ContentType:  contentType,
		Body:         objectOutput.Body,
		Size:         size,
	}, nil
}

// getRecordEntity loads one sample record entity by primary key.
func (s *serviceImpl) getRecordEntity(ctx context.Context, id int64) (*demoRecordEntity, error) {
	if err := ensureDemoRecordTableReady(ctx); err != nil {
		return nil, err
	}
	if id <= 0 {
		return nil, bizerr.NewCode(CodeDemoRecordIDRequired)
	}

	var record *demoRecordEntity
	err := tenantspi.ApplyPluginTableFilter(ctx, s.pluginTableFilter(), dao.Record.Ctx(ctx), "").
		Where(do.Record{Id: id}).
		Scan(&record)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeDemoRecordDetailQueryFailed)
	}
	if record == nil {
		return nil, bizerr.NewCode(CodeDemoRecordNotFound)
	}
	return record, nil
}

// ensureDemoRecordTableReady verifies the sample table exists before CRUD work
// continues.
func ensureDemoRecordTableReady(ctx context.Context) error {
	fields, err := dao.Record.DB().TableFields(ctx, dao.Record.Table())
	if err != nil {
		return bizerr.WrapCode(err, CodeDemoRecordTableCheckFailed)
	}
	if len(fields) == 0 {
		return bizerr.NewCode(CodeDemoRecordTableNotInstalled)
	}
	return nil
}

// normalizeListPagination applies paging defaults and max-page-size limits.
func normalizeListPagination(in *ListRecordsInput) (int, int) {
	if in == nil {
		return defaultPageNum, defaultPageSize
	}

	pageNum := in.PageNum
	if pageNum <= 0 {
		pageNum = defaultPageNum
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return pageNum, pageSize
}

// validateRecordTitle validates the required sample record title field.
func validateRecordTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return bizerr.NewCode(CodeDemoRecordTitleRequired)
	}
	return nil
}

// buildRecordListItemOutput converts one internal entity into the list item
// response shape.
func buildRecordListItemOutput(item *demoRecordEntity) *RecordListItemOutput {
	if item == nil {
		return &RecordListItemOutput{}
	}
	return &RecordListItemOutput{
		Id:             item.Id,
		Title:          item.Title,
		Content:        item.Content,
		AttachmentName: item.AttachmentName,
		HasAttachment:  strings.TrimSpace(item.AttachmentPath) != "",
		CreatedAt:      apitime.Milli(item.CreatedAt),
		UpdatedAt:      apitime.Milli(item.UpdatedAt),
	}
}

// stringPointer allocates one string pointer for optional DB mutation fields.
func stringPointer(value string) *string {
	return &value
}

// listAllAttachmentPaths returns all persisted attachment paths and tenant
// scopes stored by the sample records table.
func listAllAttachmentPaths(ctx context.Context) ([]demoAttachmentObject, error) {
	fields, err := dao.Record.DB().TableFields(ctx, dao.Record.Table())
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeDemoRecordTableCheckFailed)
	}
	if len(fields) == 0 {
		return []demoAttachmentObject{}, nil
	}

	cols := dao.Record.Columns()
	rows, err := dao.Record.Ctx(ctx).
		Fields(cols.TenantId, cols.AttachmentPath).
		All()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeDemoRecordAttachmentPathQueryFailed)
	}

	objects := make([]demoAttachmentObject, 0, len(rows))
	for _, row := range rows {
		value := strings.TrimSpace(row[cols.AttachmentPath].String())
		if value != "" {
			objects = append(objects, demoAttachmentObject{
				TenantID: row[cols.TenantId].Int(),
				Path:     value,
			})
		}
	}
	return objects, nil
}

// cleanupDemoAttachmentAfterMutationFailure removes an attachment created by a
// failed mutation and logs cleanup failures without hiding the primary error.
func (s *serviceImpl) cleanupDemoAttachmentAfterMutationFailure(ctx context.Context, attachmentPath string) {
	if cleanupErr := s.deleteDemoAttachmentFile(ctx, attachmentPath); cleanupErr != nil {
		logger.Warningf(
			ctx,
			"cleanup demo attachment after failed mutation failed path=%s err=%v",
			attachmentPath,
			cleanupErr,
		)
	}
}
