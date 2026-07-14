// Cloud object backend for Qiniu Kodo using the official Qiniu Go SDK.
//
// This file maps storagecap object operations onto Kodo upload / management
// APIs (form/resume upload, Stat/Get/Delete/List) and normalizes not-found
// responses to storagecap stable semantics.
package objstore

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/client"
	"github.com/qiniu/go-sdk/v7/storage"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/storagecap"
	settingssvc "lina-plugin-linapro-storage-qiniu/backend/internal/service/settings"
)

// qiniuBackend implements objectBackend against one Kodo bucket.
type qiniuBackend struct {
	mac      *auth.Credentials
	cfg      *storage.Config
	bucket   string
	manager  *storage.BucketManager
	uploader *storage.FormUploader
	// downloadDomain is an optional custom download host for Get.
	downloadDomain string
}

// newCloudBackend constructs a Kodo client from settings.
func newCloudBackend(snapshot *settingssvc.Snapshot) (objectBackend, error) {
	if snapshot == nil {
		return nil, bizerr.NewCode(settingssvc.CodeConfigInvalid)
	}
	var (
		accessKey  = strings.TrimSpace(snapshot.AccessKeyID)
		secretKey  = strings.TrimSpace(snapshot.SecretAccessKey)
		bucketName = strings.TrimSpace(snapshot.Bucket)
	)
	if accessKey == "" || secretKey == "" || bucketName == "" {
		return nil, bizerr.NewCode(settingssvc.CodeConfigInvalid)
	}
	mac := auth.New(accessKey, secretKey)
	cfg := storage.NewConfig()
	cfg.UseHTTPS = true

	regionID := strings.TrimSpace(snapshot.Region)
	if regionID != "" {
		region, ok := storage.GetRegionByID(storage.RegionID(normalizeRegionID(regionID)))
		if !ok {
			return nil, bizerr.NewCode(settingssvc.CodeConfigInvalid)
		}
		cfg.Region = &region
	} else {
		// Auto-resolve zone from AK + bucket when region is omitted.
		region, err := storage.GetRegion(accessKey, bucketName)
		if err != nil {
			return nil, bizerr.WrapCode(err, settingssvc.CodeConfigInvalid)
		}
		cfg.Region = region
	}

	manager := storage.NewBucketManager(mac, cfg)
	return &qiniuBackend{
		mac:            mac,
		cfg:            cfg,
		bucket:         bucketName,
		manager:        manager,
		uploader:       storage.NewFormUploader(cfg),
		downloadDomain: strings.TrimSpace(snapshot.Endpoint),
	}, nil
}

func normalizeRegionID(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "huadong", "east-china", "cn-east-1":
		return "z0"
	case "huabei", "north-china", "cn-north-1":
		return "z1"
	case "huanan", "south-china", "cn-south-1":
		return "z2"
	case "beimei", "north-america", "us-north-1":
		return "na0"
	case "xinjiapo", "singapore", "ap-southeast-1":
		return "as0"
	default:
		return raw
	}
}

func (b *qiniuBackend) uploadToken(key string, insertOnly bool) string {
	policy := storage.PutPolicy{
		Scope:   b.bucket + ":" + key,
		Expires: 3600,
	}
	if insertOnly {
		policy.InsertOnly = 1
	}
	return policy.UploadToken(b.mac)
}

func (b *qiniuBackend) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string, overwrite bool) (*objectMeta, error) {
	if !overwrite {
		_, found, err := b.Stat(ctx, key)
		if err != nil {
			return nil, err
		}
		if found {
			return nil, bizerr.NewCode(storagecap.CodeStorageObjectExists)
		}
	}
	token := b.uploadToken(key, !overwrite)
	extra := &storage.PutExtra{}
	if strings.TrimSpace(contentType) != "" {
		extra.MimeType = contentType
	}
	var ret storage.PutRet
	var err error
	if size >= 0 {
		err = b.uploader.Put(ctx, &ret, token, key, body, size, extra)
	} else {
		// Unknown size: buffer then form-upload (simpler and reliable for moderate objects).
		data, readErr := io.ReadAll(body)
		if readErr != nil {
			return nil, readErr
		}
		err = b.uploader.Put(ctx, &ret, token, key, bytes.NewReader(data), int64(len(data)), extra)
		size = int64(len(data))
	}
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	etag := ret.Hash
	return &objectMeta{Key: key, Size: size, ContentType: contentType, ETag: etag, UpdatedAt: &now}, nil
}

func (b *qiniuBackend) Get(ctx context.Context, key string) (*objectMeta, io.ReadCloser, bool, error) {
	input := &storage.GetObjectInput{
		Context:    ctx,
		PresignUrl: true,
	}
	if b.downloadDomain != "" {
		input.DownloadDomains = []string{b.downloadDomain}
	}
	output, err := b.manager.Get(b.bucket, key, input)
	if err != nil {
		// Get may return partial output + error; close body if present.
		if output != nil && output.Body != nil {
			_ = output.Body.Close()
		}
		if isQiniuNotFound(err) {
			return nil, nil, false, nil
		}
		return nil, nil, false, err
	}
	if output == nil || output.Body == nil {
		return nil, nil, false, nil
	}
	meta := &objectMeta{
		Key:         key,
		Size:        output.ContentLength,
		ContentType: output.ContentType,
		ETag:        strings.Trim(output.ETag, `"`),
	}
	if !output.LastModified.IsZero() {
		t := output.LastModified.UTC()
		meta.UpdatedAt = &t
	}
	// Prefer Stat size/etag when ContentLength is unknown.
	if meta.Size < 0 {
		if st, found, stErr := b.Stat(ctx, key); stErr == nil && found && st != nil {
			meta.Size = st.Size
			if meta.ETag == "" {
				meta.ETag = st.ETag
			}
			if meta.ContentType == "" {
				meta.ContentType = st.ContentType
			}
			if meta.UpdatedAt == nil {
				meta.UpdatedAt = st.UpdatedAt
			}
		}
	}
	return meta, output.Body, true, nil
}

func (b *qiniuBackend) Delete(ctx context.Context, key string) error {
	_ = ctx
	err := b.manager.Delete(b.bucket, key)
	if err != nil && isQiniuNotFound(err) {
		return nil
	}
	return err
}

func (b *qiniuBackend) List(ctx context.Context, prefix string, cursor string, limit int) ([]*objectMeta, string, error) {
	_ = ctx
	if limit <= 0 {
		limit = storagecap.DefaultListLimit
	}
	if limit > 1000 {
		limit = 1000
	}
	entries, _, nextMarker, hasNext, err := b.manager.ListFiles(b.bucket, prefix, "", cursor, limit)
	if err != nil {
		return nil, "", err
	}
	metas := make([]*objectMeta, 0, len(entries))
	for _, item := range entries {
		if item.IsEmpty() {
			continue
		}
		meta := &objectMeta{
			Key:         item.Key,
			Size:        item.Fsize,
			ContentType: item.MimeType,
			ETag:        item.Hash,
		}
		if item.PutTime > 0 {
			// PutTime unit is 100ns; divide by 1e7 for Unix seconds.
			t := time.Unix(item.PutTime/1e7, 0).UTC()
			meta.UpdatedAt = &t
		}
		metas = append(metas, meta)
	}
	next := ""
	if hasNext {
		next = nextMarker
	}
	return metas, next, nil
}

func (b *qiniuBackend) Stat(ctx context.Context, key string) (*objectMeta, bool, error) {
	_ = ctx
	info, err := b.manager.Stat(b.bucket, key)
	if err != nil {
		if isQiniuNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	meta := &objectMeta{
		Key:         key,
		Size:        info.Fsize,
		ContentType: info.MimeType,
		ETag:        info.Hash,
	}
	if info.PutTime > 0 {
		t := time.Unix(info.PutTime/1e7, 0).UTC()
		meta.UpdatedAt = &t
	}
	return meta, true, nil
}

func (b *qiniuBackend) HeadBucket(ctx context.Context) error {
	_ = ctx
	// Cheap probe: list at most one object under the configured bucket.
	entries, commonPrefixes, nextMarker, hasNext, err := b.manager.ListFiles(b.bucket, "", "", "", 1)
	if err != nil {
		return err
	}
	// Connectivity is proven by a successful list; listing payload is intentionally not returned.
	if len(entries) == 0 && len(commonPrefixes) == 0 && nextMarker == "" && !hasNext {
		return nil
	}
	return nil
}

func isQiniuNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, storage.ErrNoSuchFile) {
		return true
	}
	var ei *client.ErrorInfo
	if errors.As(err, &ei) {
		// 612: no such file or directory; 631: no such bucket treated as config error not not-found for Stat path.
		if ei.Code == 612 {
			return true
		}
		msg := strings.ToLower(ei.Err)
		if strings.Contains(msg, "no such file") || strings.Contains(msg, "not found") {
			return true
		}
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "no such file") || strings.Contains(msg, "612")
}
