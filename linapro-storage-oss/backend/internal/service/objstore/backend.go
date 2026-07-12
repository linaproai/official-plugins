package objstore

import (
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
