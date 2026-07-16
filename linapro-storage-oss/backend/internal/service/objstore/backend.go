package objstore

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/storagecap"
	settingssvc "lina-plugin-linapro-storage-oss/backend/internal/service/settings"
)

type ossBackend struct {
	bucket *oss.Bucket
}

func newCloudBackend(snapshot *settingssvc.Snapshot) (objectBackend, error) {
	if snapshot == nil {
		return nil, bizerr.NewCode(settingssvc.CodeConfigInvalid)
	}
	region := strings.TrimSpace(snapshot.Region)
	bucketName := strings.TrimSpace(snapshot.Bucket)
	if region == "" || bucketName == "" {
		return nil, bizerr.NewCode(settingssvc.CodeConfigInvalid)
	}
	endpoint := strings.TrimSpace(snapshot.Endpoint)
	if endpoint == "" {
		endpoint = "https://oss-" + region + ".aliyuncs.com"
	}
	client, err := oss.New(endpoint, strings.TrimSpace(snapshot.AccessKeyID), strings.TrimSpace(snapshot.SecretAccessKey))
	if err != nil {
		return nil, bizerr.WrapCode(err, settingssvc.CodeConfigInvalid)
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, bizerr.WrapCode(err, settingssvc.CodeConfigInvalid)
	}
	return &ossBackend{bucket: bucket}, nil
}

func (b *ossBackend) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string, overwrite bool) (*objectMeta, error) {
	_ = ctx
	if !overwrite {
		exists, err := b.bucket.IsObjectExist(key)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, bizerr.NewCode(storagecap.CodeStorageObjectExists)
		}
	}
	options := []oss.Option{}
	if strings.TrimSpace(contentType) != "" {
		options = append(options, oss.ContentType(contentType))
	}
	if err := b.bucket.PutObject(key, body, options...); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	return &objectMeta{Key: key, Size: size, ContentType: contentType, UpdatedAt: &now}, nil
}

func (b *ossBackend) Get(ctx context.Context, key string) (*objectMeta, io.ReadCloser, bool, error) {
	_ = ctx
	body, err := b.bucket.GetObject(key)
	if err != nil {
		if isOSSNotFound(err) {
			return nil, nil, false, nil
		}
		return nil, nil, false, err
	}
	meta, _, err := b.Stat(ctx, key)
	if err != nil {
		_ = body.Close()
		return nil, nil, false, err
	}
	if meta == nil {
		meta = &objectMeta{Key: key}
	}
	return meta, body, true, nil
}

func (b *ossBackend) Delete(ctx context.Context, key string) error {
	_ = ctx
	err := b.bucket.DeleteObject(key)
	if err != nil && isOSSNotFound(err) {
		return nil
	}
	return err
}

func (b *ossBackend) List(ctx context.Context, prefix string, cursor string, limit int) ([]*objectMeta, string, error) {
	_ = ctx
	options := []oss.Option{oss.Prefix(prefix), oss.MaxKeys(limit)}
	if cursor != "" {
		options = append(options, oss.Marker(cursor))
	}
	result, err := b.bucket.ListObjects(options...)
	if err != nil {
		return nil, "", err
	}
	metas := make([]*objectMeta, 0, len(result.Objects))
	for _, item := range result.Objects {
		meta := &objectMeta{Key: item.Key, Size: item.Size, ETag: strings.Trim(item.ETag, `"`)}
		if !item.LastModified.IsZero() {
			t := item.LastModified.UTC()
			meta.UpdatedAt = &t
		}
		metas = append(metas, meta)
	}
	next := ""
	if result.IsTruncated && len(metas) > 0 {
		next = metas[len(metas)-1].Key
	}
	return metas, next, nil
}

func (b *ossBackend) Stat(ctx context.Context, key string) (*objectMeta, bool, error) {
	_ = ctx
	header, err := b.bucket.GetObjectMeta(key)
	if err != nil {
		if isOSSNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	meta := &objectMeta{Key: key, ContentType: header.Get("Content-Type"), ETag: strings.Trim(header.Get("ETag"), `"`)}
	if cl := header.Get("Content-Length"); cl != "" {
		if size, err := strconv.ParseInt(cl, 10, 64); err == nil {
			meta.Size = size
		}
	}
	if lm := header.Get("Last-Modified"); lm != "" {
		if t, err := time.Parse(time.RFC1123, lm); err == nil {
			t = t.UTC()
			meta.UpdatedAt = &t
		}
	}
	return meta, true, nil
}

func (b *ossBackend) HeadBucket(ctx context.Context) error {
	_ = ctx
	// List with max 1 is a cheap probe when dedicated HeadBucket is awkward.
	_, err := b.bucket.ListObjects(oss.MaxKeys(1))
	return err
}

func normalizePresignTTL(ttl time.Duration) time.Duration {
	if ttl <= 0 {
		ttl = time.Hour
	}
	if ttl > time.Hour {
		ttl = time.Hour
	}
	return ttl
}

func (b *ossBackend) PresignPut(_ context.Context, key string, contentType string, ttl time.Duration) (string, map[string]string, time.Time, error) {
	ttl = normalizePresignTTL(ttl)
	options := []oss.Option{}
	headers := map[string]string{}
	if strings.TrimSpace(contentType) != "" {
		options = append(options, oss.ContentType(contentType))
		headers["Content-Type"] = contentType
	}
	signed, err := b.bucket.SignURL(key, oss.HTTPPut, int64(ttl.Seconds()), options...)
	if err != nil {
		return "", nil, time.Time{}, err
	}
	return signed, headers, time.Now().UTC().Add(ttl), nil
}

func (b *ossBackend) PresignGet(_ context.Context, key string, ttl time.Duration) (string, time.Time, error) {
	ttl = normalizePresignTTL(ttl)
	signed, err := b.bucket.SignURL(key, oss.HTTPGet, int64(ttl.Seconds()))
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, time.Now().UTC().Add(ttl), nil
}

func isOSSNotFound(err error) bool {
	if err == nil {
		return false
	}
	var serr oss.ServiceError
	if errors.As(err, &serr) {
		return serr.StatusCode == http.StatusNotFound || serr.Code == "NoSuchKey" || serr.Code == "NotFound"
	}
	return strings.Contains(err.Error(), "NoSuchKey") || strings.Contains(err.Error(), "404")
}

func (b *ossBackend) CreateMultipart(_ context.Context, key string, contentType string) (string, error) {
	options := []oss.Option{}
	if strings.TrimSpace(contentType) != "" {
		options = append(options, oss.ContentType(contentType))
	}
	imur, err := b.bucket.InitiateMultipartUpload(key, options...)
	if err != nil {
		return "", err
	}
	return imur.UploadID, nil
}

func (b *ossBackend) UploadPart(_ context.Context, key string, uploadID string, partNumber int32, body io.Reader, size int64) (string, error) {
	if size < 0 {
		data, err := io.ReadAll(body)
		if err != nil {
			return "", err
		}
		body = bytes.NewReader(data)
		size = int64(len(data))
	}
	imur := oss.InitiateMultipartUploadResult{Key: key, UploadID: uploadID, Bucket: b.bucket.BucketName}
	part, err := b.bucket.UploadPart(imur, body, size, int(partNumber))
	if err != nil {
		return "", err
	}
	return strings.Trim(part.ETag, `"`), nil
}

func (b *ossBackend) CompleteMultipart(_ context.Context, key string, uploadID string, parts []completedPart) (*objectMeta, error) {
	imur := oss.InitiateMultipartUploadResult{Key: key, UploadID: uploadID, Bucket: b.bucket.BucketName}
	uploadParts := make([]oss.UploadPart, 0, len(parts))
	for _, part := range parts {
		uploadParts = append(uploadParts, oss.UploadPart{PartNumber: int(part.PartNumber), ETag: part.ETag})
	}
	_, err := b.bucket.CompleteMultipartUpload(imur, uploadParts)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	return &objectMeta{Key: key, UpdatedAt: &now}, nil
}

func (b *ossBackend) AbortMultipart(_ context.Context, key string, uploadID string) error {
	imur := oss.InitiateMultipartUploadResult{Key: key, UploadID: uploadID, Bucket: b.bucket.BucketName}
	return b.bucket.AbortMultipartUpload(imur)
}

func (b *ossBackend) PresignUploadPart(_ context.Context, key string, uploadID string, partNumber int32, ttl time.Duration) (string, map[string]string, time.Time, error) {
	ttl = normalizePresignTTL(ttl)
	signed, err := b.bucket.SignURL(key, oss.HTTPPut, int64(ttl.Seconds()),
		oss.AddParam("partNumber", strconv.Itoa(int(partNumber))),
		oss.AddParam("uploadId", uploadID),
	)
	if err != nil {
		return "", nil, time.Time{}, err
	}
	return signed, map[string]string{}, time.Now().UTC().Add(ttl), nil
}

