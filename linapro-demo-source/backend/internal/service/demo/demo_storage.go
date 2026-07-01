// demo_storage.go implements plugin-owned attachment storage helpers and
// uninstall-time cleanup through the host Storage() capability.

package demo

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/grand"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/storagecap"
)

// Attachment storage constants define the storage namespace and default upload
// ceiling used by the sample source plugin.
const (
	demoAttachmentStoragePrefix = "demo-record-files"
	demoAttachmentMaxBytes      = 32 * 1024 * 1024
	defaultUploadMaxSizeMB      = demoAttachmentMaxBytes / 1024 / 1024
	maxStoragePurgeListLoops    = 100
)

// demoAttachmentObject identifies one stored attachment and its tenant scope.
type demoAttachmentObject struct {
	// TenantID is the tenant scope used by Storage() for this object.
	TenantID int
	// Path is the plugin logical object path.
	Path string
}

// PurgeStorageData clears plugin-owned attachment files before uninstall SQL drops the data table.
func PurgeStorageData(ctx context.Context, storageSvc storagecap.Service) error {
	if storageSvc == nil {
		return bizerr.NewCode(CodeDemoAttachmentStorageUnavailable)
	}
	objects, err := listAllAttachmentPaths(ctx)
	if err != nil {
		return err
	}
	tenantIDs := map[int]struct{}{0: {}}
	for _, object := range objects {
		tenantIDs[object.TenantID] = struct{}{}
		objectCtx := demoAttachmentTenantContext(ctx, object.TenantID)
		if err = deleteDemoAttachmentFile(objectCtx, storageSvc, object.Path); err != nil {
			return err
		}
	}

	for tenantID := range tenantIDs {
		objectCtx := demoAttachmentTenantContext(ctx, tenantID)
		if err = purgeDemoAttachmentPrefix(ctx, storageSvc, objectCtx); err != nil {
			return err
		}
	}
	return nil
}

// PurgeStorageData delegates service cleanup to the injected plugin storage service.
func (s *serviceImpl) PurgeStorageData(ctx context.Context) error {
	return PurgeStorageData(ctx, s.storageSvc)
}

// saveDemoAttachmentFile stores one optional uploaded attachment into the
// plugin-owned storage area.
func (s *serviceImpl) saveDemoAttachmentFile(
	ctx context.Context,
	file *ghttp.UploadFile,
) (originalName string, relativePath string, err error) {
	if file == nil {
		return "", "", nil
	}
	if s.storageSvc == nil {
		return "", "", bizerr.NewCode(CodeDemoAttachmentStorageUnavailable)
	}
	if err = validateDemoAttachmentFileSize(file); err != nil {
		return "", "", err
	}

	sanitizedName := sanitizeAttachmentFilename(file.Filename)
	source, err := file.Open()
	if err != nil {
		return "", "", bizerr.WrapCode(err, CodeDemoAttachmentOpenFailed)
	}
	defer func() {
		closeErr := source.Close()
		if err == nil && closeErr != nil {
			err = bizerr.WrapCode(closeErr, CodeDemoAttachmentSourceCloseFailed)
		}
	}()

	var (
		now        = gtime.Now()
		ext        = strings.TrimPrefix(filepath.Ext(sanitizedName), ".")
		storedName = fmt.Sprintf("%s_%s", now.Format("Ymd_His"), grand.S(8))
	)
	if ext != "" {
		storedName += "." + gstr.ToLower(ext)
	}

	objectPath := path.Join(demoAttachmentStoragePrefix, now.Format("Y"), now.Format("m"), storedName)
	if _, err = s.storageSvc.Put(ctx, storagecap.PutInput{
		Path:        objectPath,
		Body:        source,
		Size:        file.Size,
		ContentType: detectDemoAttachmentContentType(sanitizedName),
		Overwrite:   false,
	}); err != nil {
		return "", "", bizerr.WrapCode(err, CodeDemoAttachmentWriteFailed)
	}

	return sanitizedName, objectPath, nil
}

// deleteDemoAttachmentFile removes one stored attachment when its relative path
// is present.
func (s *serviceImpl) deleteDemoAttachmentFile(ctx context.Context, relativePath string) error {
	return deleteDemoAttachmentFile(ctx, s.storageSvc, relativePath)
}

// deleteDemoAttachmentFile removes one plugin storage object through the given service.
func deleteDemoAttachmentFile(ctx context.Context, storageSvc storagecap.Service, relativePath string) error {
	if strings.TrimSpace(relativePath) == "" {
		return nil
	}
	if storageSvc == nil {
		return bizerr.NewCode(CodeDemoAttachmentStorageUnavailable)
	}
	if err := storageSvc.Delete(ctx, storagecap.DeleteInput{Path: relativePath}); err != nil {
		return bizerr.WrapCode(err, CodeDemoAttachmentDeleteFailed)
	}
	return nil
}

// purgeDemoAttachmentPrefix deletes remaining objects under the attachment prefix.
func purgeDemoAttachmentPrefix(logCtx context.Context, storageSvc storagecap.Service, objectCtx context.Context) error {
	for loop := 0; loop < maxStoragePurgeListLoops; loop++ {
		out, listErr := storageSvc.List(objectCtx, storagecap.ListInput{
			Prefix: demoAttachmentStoragePrefix + "/",
			Limit:  storagecap.MaxListLimit,
		})
		if listErr != nil {
			return bizerr.WrapCode(listErr, CodeDemoAttachmentStoragePurgeFailed)
		}
		if out == nil || len(out.Objects) == 0 {
			return nil
		}
		for _, object := range out.Objects {
			if object == nil {
				continue
			}
			if err := deleteDemoAttachmentFile(objectCtx, storageSvc, object.Path); err != nil {
				logger.Warningf(
					logCtx,
					"purge demo attachment prefix failed path=%s err=%v",
					object.Path,
					err,
				)
				return err
			}
		}
		if len(out.Objects) < storagecap.MaxListLimit {
			return nil
		}
	}
	return bizerr.NewCode(CodeDemoAttachmentStoragePurgeFailed)
}

// demoAttachmentTenantContext binds the tenant scope used by Storage() object keys.
func demoAttachmentTenantContext(ctx context.Context, tenantID int) context.Context {
	current := bizctxcap.CurrentFromContext(ctx)
	if tenantID <= 0 {
		current.TenantID = 0
		current.PlatformBypass = true
	} else {
		current.TenantID = tenantID
		current.PlatformBypass = false
	}
	return bizctxcap.WithCurrentContext(ctx, current)
}

// validateDemoAttachmentFileSize enforces the source-plugin demo attachment ceiling.
func validateDemoAttachmentFileSize(file *ghttp.UploadFile) error {
	if file == nil {
		return nil
	}
	if file.Size > demoAttachmentMaxBytes {
		return bizerr.NewCode(CodeDemoAttachmentSizeTooLarge, bizerr.P("maxSizeMB", defaultUploadMaxSizeMB))
	}
	return nil
}

// detectDemoAttachmentContentType derives a stable MIME type from the upload name.
func detectDemoAttachmentContentType(filename string) string {
	contentType := mime.TypeByExtension(filepath.Ext(filename))
	if strings.TrimSpace(contentType) != "" {
		return contentType
	}
	return http.DetectContentType(nil)
}

// sanitizeAttachmentFilename strips unsafe characters and truncates overly long
// attachment filenames.
func sanitizeAttachmentFilename(filename string) string {
	filename = filepath.Base(filename)
	filename = strings.ReplaceAll(filename, "\x00", "")
	if strings.TrimSpace(filename) == "" {
		return "attachment"
	}

	disallowed := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, item := range disallowed {
		filename = strings.ReplaceAll(filename, item, "_")
	}
	if len(filename) > 255 {
		ext := filepath.Ext(filename)
		name := filename[:255-len(ext)]
		filename = name + ext
	}
	return filename
}
