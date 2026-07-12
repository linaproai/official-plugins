// Cloud object backend for Huawei Cloud OBS using the official OBS Go SDK.
//
// This file maps storagecap object operations onto OBS bucket APIs and
// normalizes not-found / conflict responses to storagecap stable error codes.
package objstore

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/storagecap"
	settingssvc "lina-plugin-linapro-storage-obs/backend/internal/service/settings"
)

// obsBackend implements objectBackend against a single OBS bucket.
type obsBackend struct {
	client *obs.ObsClient
	bucket string
}

// newCloudBackend constructs an OBS client from settings.
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
		endpoint = "https://obs." + region + ".myhuaweicloud.com"
	}
	client, err := obs.New(
		strings.TrimSpace(snapshot.AccessKeyID),
		strings.TrimSpace(snapshot.SecretAccessKey),
		endpoint,
	)
	if err != nil {
		return nil, bizerr.WrapCode(err, settingssvc.CodeConfigInvalid)
	}
	return &obsBackend{client: client, bucket: bucketName}, nil
}

func (b *obsBackend) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string, overwrite bool) (*objectMeta, error) {
	_ = ctx
	if !overwrite {
		_, found, err := b.Stat(ctx, key)
		if err != nil {
			return nil, err
		}
		if found {
			return nil, bizerr.NewCode(storagecap.CodeStorageObjectExists)
		}
	}
	input := &obs.PutObjectInput{}
	input.Bucket = b.bucket
	input.Key = key
	input.Body = body
	if size >= 0 {
		input.ContentLength = size
	}
	if strings.TrimSpace(contentType) != "" {
		input.ContentType = contentType
	}
	output, err := b.client.PutObject(input)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	etag := ""
	if output != nil {
		etag = strings.Trim(output.ETag, `"`)
	}
	return &objectMeta{Key: key, Size: size, ContentType: contentType, ETag: etag, UpdatedAt: &now}, nil
}

func (b *obsBackend) Get(ctx context.Context, key string) (*objectMeta, io.ReadCloser, bool, error) {
	_ = ctx
	output, err := b.client.GetObject(&obs.GetObjectInput{
		GetObjectMetadataInput: obs.GetObjectMetadataInput{
			Bucket: b.bucket,
			Key:    key,
		},
	})
	if err != nil {
		if isOBSNotFound(err) {
			return nil, nil, false, nil
		}
		return nil, nil, false, err
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
	return meta, output.Body, true, nil
}

func (b *obsBackend) Delete(ctx context.Context, key string) error {
	_ = ctx
	_, err := b.client.DeleteObject(&obs.DeleteObjectInput{
		Bucket: b.bucket,
		Key:    key,
	})
	if err != nil && isOBSNotFound(err) {
		return nil
	}
	return err
}

func (b *obsBackend) List(ctx context.Context, prefix string, cursor string, limit int) ([]*objectMeta, string, error) {
	_ = ctx
	input := &obs.ListObjectsInput{}
	input.Bucket = b.bucket
	input.Prefix = prefix
	input.MaxKeys = limit
	if cursor != "" {
		input.Marker = cursor
	}
	result, err := b.client.ListObjects(input)
	if err != nil {
		return nil, "", err
	}
	metas := make([]*objectMeta, 0, len(result.Contents))
	for _, item := range result.Contents {
		meta := &objectMeta{
			Key:  item.Key,
			Size: item.Size,
			ETag: strings.Trim(item.ETag, `"`),
		}
		if !item.LastModified.IsZero() {
			t := item.LastModified.UTC()
			meta.UpdatedAt = &t
		}
		metas = append(metas, meta)
	}
	next := ""
	if result.IsTruncated {
		if result.NextMarker != "" {
			next = result.NextMarker
		} else if len(metas) > 0 {
			next = metas[len(metas)-1].Key
		}
	}
	return metas, next, nil
}

func (b *obsBackend) Stat(ctx context.Context, key string) (*objectMeta, bool, error) {
	_ = ctx
	output, err := b.client.GetObjectMetadata(&obs.GetObjectMetadataInput{
		Bucket: b.bucket,
		Key:    key,
	})
	if err != nil {
		if isOBSNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
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
	return meta, true, nil
}

func (b *obsBackend) HeadBucket(ctx context.Context) error {
	_ = ctx
	_, err := b.client.HeadBucket(b.bucket)
	return err
}

func isOBSNotFound(err error) bool {
	if err == nil {
		return false
	}
	var oerr obs.ObsError
	if errors.As(err, &oerr) {
		if oerr.StatusCode == http.StatusNotFound {
			return true
		}
		code := strings.ToLower(oerr.Code)
		return code == "nosuchkey" || code == "notfound" || code == "nosuchbucket"
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "nosuchkey") || strings.Contains(msg, "404")
}
