// Cloud object backend for Huawei Cloud OBS using the official OBS Go SDK.
//
// This file maps storagecap object operations onto OBS bucket APIs and
// normalizes not-found / conflict responses to storagecap stable error codes.
package objstore

import (
	"context"
	"fmt"
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

func normalizePresignTTL(ttl time.Duration) time.Duration {
	if ttl <= 0 {
		ttl = time.Hour
	}
	if ttl > time.Hour {
		ttl = time.Hour
	}
	return ttl
}

func (b *obsBackend) PresignPut(_ context.Context, key string, contentType string, ttl time.Duration) (string, map[string]string, time.Time, error) {
	ttl = normalizePresignTTL(ttl)
	input := &obs.CreateSignedUrlInput{
		Method:  obs.HttpMethodPut,
		Bucket:  b.bucket,
		Key:     key,
		Expires: int(ttl.Seconds()),
	}
	headers := map[string]string{}
	if strings.TrimSpace(contentType) != "" {
		input.Headers = map[string]string{"Content-Type": contentType}
		headers["Content-Type"] = contentType
	}
	out, err := b.client.CreateSignedUrl(input)
	if err != nil {
		return "", nil, time.Time{}, err
	}
	// Prefer signed headers returned by the SDK when present.
	if out != nil && out.ActualSignedRequestHeaders != nil {
		for k, vals := range out.ActualSignedRequestHeaders {
			if len(vals) > 0 && strings.TrimSpace(vals[0]) != "" {
				headers[k] = vals[0]
			}
		}
	}
	return out.SignedUrl, headers, time.Now().UTC().Add(ttl), nil
}

func (b *obsBackend) PresignGet(_ context.Context, key string, ttl time.Duration) (string, time.Time, error) {
	ttl = normalizePresignTTL(ttl)
	out, err := b.client.CreateSignedUrl(&obs.CreateSignedUrlInput{
		Method:  obs.HttpMethodGet,
		Bucket:  b.bucket,
		Key:     key,
		Expires: int(ttl.Seconds()),
	})
	if err != nil {
		return "", time.Time{}, err
	}
	return out.SignedUrl, time.Now().UTC().Add(ttl), nil
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

func (b *obsBackend) CreateMultipart(_ context.Context, key string, contentType string) (string, error) {
	input := &obs.InitiateMultipartUploadInput{}
	input.Bucket = b.bucket
	input.Key = key
	if strings.TrimSpace(contentType) != "" {
		input.ContentType = contentType
	}
	out, err := b.client.InitiateMultipartUpload(input)
	if err != nil {
		return "", err
	}
	if out == nil || strings.TrimSpace(out.UploadId) == "" {
		return "", fmt.Errorf("obs create multipart returned empty upload id")
	}
	return out.UploadId, nil
}

func (b *obsBackend) UploadPart(_ context.Context, key string, uploadID string, partNumber int32, body io.Reader, size int64) (string, error) {
	input := &obs.UploadPartInput{
		Bucket:     b.bucket,
		Key:        key,
		UploadId:   uploadID,
		PartNumber: int(partNumber),
		Body:       body,
	}
	if size >= 0 {
		input.PartSize = size
	}
	out, err := b.client.UploadPart(input)
	if err != nil {
		return "", err
	}
	if out == nil {
		return "", nil
	}
	return strings.Trim(out.ETag, `"`), nil
}

func (b *obsBackend) CompleteMultipart(_ context.Context, key string, uploadID string, parts []completedPart) (*objectMeta, error) {
	obsParts := make([]obs.Part, 0, len(parts))
	for _, part := range parts {
		obsParts = append(obsParts, obs.Part{PartNumber: int(part.PartNumber), ETag: part.ETag})
	}
	input := &obs.CompleteMultipartUploadInput{
		Bucket:   b.bucket,
		Key:      key,
		UploadId: uploadID,
		Parts:    obsParts,
	}
	out, err := b.client.CompleteMultipartUpload(input)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	meta := &objectMeta{Key: key, UpdatedAt: &now}
	if out != nil {
		meta.ETag = strings.Trim(out.ETag, `"`)
	}
	return meta, nil
}

func (b *obsBackend) AbortMultipart(_ context.Context, key string, uploadID string) error {
	_, err := b.client.AbortMultipartUpload(&obs.AbortMultipartUploadInput{
		Bucket:   b.bucket,
		Key:      key,
		UploadId: uploadID,
	})
	return err
}

func (b *obsBackend) PresignUploadPart(_ context.Context, key string, uploadID string, partNumber int32, ttl time.Duration) (string, map[string]string, time.Time, error) {
	ttl = normalizePresignTTL(ttl)
	out, err := b.client.CreateSignedUrl(&obs.CreateSignedUrlInput{
		Method:  obs.HttpMethodPut,
		Bucket:  b.bucket,
		Key:     key,
		Expires: int(ttl.Seconds()),
		QueryParams: map[string]string{
			"partNumber": fmt.Sprintf("%d", partNumber),
			"uploadId":   uploadID,
		},
	})
	if err != nil {
		return "", nil, time.Time{}, err
	}
	headers := map[string]string{}
	if out != nil && out.ActualSignedRequestHeaders != nil {
		for k, vals := range out.ActualSignedRequestHeaders {
			if len(vals) > 0 && strings.TrimSpace(vals[0]) != "" {
				headers[k] = vals[0]
			}
		}
	}
	return out.SignedUrl, headers, time.Now().UTC().Add(ttl), nil
}

