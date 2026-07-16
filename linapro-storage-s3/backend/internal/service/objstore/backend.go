package objstore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/storagecap"
	settingssvc "lina-plugin-linapro-storage-s3/backend/internal/service/settings"
)

type s3Backend struct {
	client *s3.Client
	bucket string
}

func newCloudBackend(snapshot *settingssvc.Snapshot) (objectBackend, error) {
	if snapshot == nil {
		return nil, bizerr.NewCode(settingssvc.CodeConfigInvalid)
	}
	endpoint := strings.TrimSpace(snapshot.Endpoint)
	bucket := strings.TrimSpace(snapshot.Bucket)
	if endpoint == "" || bucket == "" {
		return nil, bizerr.NewCode(settingssvc.CodeConfigInvalid)
	}
	cfg := aws.Config{
		Region: settingssvc.EffectiveRegion(snapshot.Region),
		Credentials: credentials.NewStaticCredentialsProvider(
			strings.TrimSpace(snapshot.AccessKeyID),
			strings.TrimSpace(snapshot.SecretAccessKey),
			"",
		),
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = snapshot.ForcePathStyle
	})
	return &s3Backend{client: client, bucket: bucket}, nil
}

func (b *s3Backend) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string, overwrite bool) (*objectMeta, error) {
	if !overwrite {
		_, found, err := b.Stat(ctx, key)
		if err != nil {
			return nil, err
		}
		if found {
			return nil, bizerr.NewCode(storagecap.CodeStorageObjectExists)
		}
	}
	input := &s3.PutObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
		Body:   body,
	}
	if strings.TrimSpace(contentType) != "" {
		input.ContentType = aws.String(contentType)
	}
	if size >= 0 {
		input.ContentLength = aws.Int64(size)
	}
	out, err := b.client.PutObject(ctx, input)
	if err != nil {
		return nil, mapS3Err(err)
	}
	now := time.Now().UTC()
	etag := ""
	if out.ETag != nil {
		etag = strings.Trim(*out.ETag, `"`)
	}
	return &objectMeta{Key: key, Size: size, ContentType: contentType, ETag: etag, UpdatedAt: &now}, nil
}

func (b *s3Backend) Get(ctx context.Context, key string) (*objectMeta, io.ReadCloser, bool, error) {
	out, err := b.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(b.bucket), Key: aws.String(key)})
	if err != nil {
		if isNotFound(err) {
			return nil, nil, false, nil
		}
		return nil, nil, false, mapS3Err(err)
	}
	meta := &objectMeta{Key: key}
	if out.ContentLength != nil {
		meta.Size = *out.ContentLength
	}
	if out.ContentType != nil {
		meta.ContentType = *out.ContentType
	}
	if out.ETag != nil {
		meta.ETag = strings.Trim(*out.ETag, `"`)
	}
	if out.LastModified != nil {
		t := out.LastModified.UTC()
		meta.UpdatedAt = &t
	}
	return meta, out.Body, true, nil
}

func (b *s3Backend) Delete(ctx context.Context, key string) error {
	_, err := b.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(b.bucket), Key: aws.String(key)})
	return mapS3Err(err)
}

func (b *s3Backend) List(ctx context.Context, prefix string, cursor string, limit int) ([]*objectMeta, string, error) {
	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(b.bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int32(int32(limit)),
	}
	if cursor != "" {
		input.StartAfter = aws.String(cursor)
	}
	out, err := b.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, "", mapS3Err(err)
	}
	metas := make([]*objectMeta, 0, len(out.Contents))
	for _, item := range out.Contents {
		if item.Key == nil {
			continue
		}
		meta := &objectMeta{Key: *item.Key}
		if item.Size != nil {
			meta.Size = *item.Size
		}
		if item.ETag != nil {
			meta.ETag = strings.Trim(*item.ETag, `"`)
		}
		if item.LastModified != nil {
			t := item.LastModified.UTC()
			meta.UpdatedAt = &t
		}
		metas = append(metas, meta)
	}
	next := ""
	if aws.ToBool(out.IsTruncated) && len(metas) > 0 {
		next = metas[len(metas)-1].Key
	}
	return metas, next, nil
}

func (b *s3Backend) Stat(ctx context.Context, key string) (*objectMeta, bool, error) {
	out, err := b.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(b.bucket), Key: aws.String(key)})
	if err != nil {
		if isNotFound(err) {
			return nil, false, nil
		}
		return nil, false, mapS3Err(err)
	}
	meta := &objectMeta{Key: key}
	if out.ContentLength != nil {
		meta.Size = *out.ContentLength
	}
	if out.ContentType != nil {
		meta.ContentType = *out.ContentType
	}
	if out.ETag != nil {
		meta.ETag = strings.Trim(*out.ETag, `"`)
	}
	if out.LastModified != nil {
		t := out.LastModified.UTC()
		meta.UpdatedAt = &t
	}
	return meta, true, nil
}

func (b *s3Backend) HeadBucket(ctx context.Context) error {
	_, err := b.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(b.bucket)})
	return mapS3Err(err)
}

func (b *s3Backend) PresignPut(ctx context.Context, key string, contentType string, ttl time.Duration) (string, map[string]string, time.Time, error) {
	if ttl <= 0 {
		ttl = time.Hour
	}
	if ttl > time.Hour {
		ttl = time.Hour
	}
	presigner := s3.NewPresignClient(b.client)
	input := &s3.PutObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	}
	headers := map[string]string{}
	if strings.TrimSpace(contentType) != "" {
		input.ContentType = aws.String(contentType)
		headers["Content-Type"] = contentType
	}
	out, err := presigner.PresignPutObject(ctx, input, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", nil, time.Time{}, mapS3Err(err)
	}
	expiresAt := time.Now().UTC().Add(ttl)
	return out.URL, headers, expiresAt, nil
}

func (b *s3Backend) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, time.Time, error) {
	if ttl <= 0 {
		ttl = time.Hour
	}
	if ttl > time.Hour {
		ttl = time.Hour
	}
	presigner := s3.NewPresignClient(b.client)
	out, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", time.Time{}, mapS3Err(err)
	}
	return out.URL, time.Now().UTC().Add(ttl), nil
}

func (b *s3Backend) CreateMultipart(ctx context.Context, key string, contentType string) (string, error) {
	input := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	}
	if strings.TrimSpace(contentType) != "" {
		input.ContentType = aws.String(contentType)
	}
	out, err := b.client.CreateMultipartUpload(ctx, input)
	if err != nil {
		return "", mapS3Err(err)
	}
	if out == nil || out.UploadId == nil {
		return "", fmt.Errorf("s3 create multipart returned empty upload id")
	}
	return *out.UploadId, nil
}

func (b *s3Backend) UploadPart(ctx context.Context, key string, uploadID string, partNumber int32, body io.Reader, size int64) (string, error) {
	input := &s3.UploadPartInput{
		Bucket:     aws.String(b.bucket),
		Key:        aws.String(key),
		UploadId:   aws.String(uploadID),
		PartNumber: aws.Int32(partNumber),
		Body:       body,
	}
	if size >= 0 {
		input.ContentLength = aws.Int64(size)
	}
	out, err := b.client.UploadPart(ctx, input)
	if err != nil {
		return "", mapS3Err(err)
	}
	if out == nil || out.ETag == nil {
		return "", nil
	}
	return strings.Trim(*out.ETag, `"`), nil
}

func (b *s3Backend) CompleteMultipart(ctx context.Context, key string, uploadID string, parts []completedPart) (*objectMeta, error) {
	completed := make([]types.CompletedPart, 0, len(parts))
	for _, part := range parts {
		etag := part.ETag
		if etag != "" && !strings.HasPrefix(etag, `"`) {
			etag = `"` + etag + `"`
		}
		completed = append(completed, types.CompletedPart{
			ETag:       aws.String(etag),
			PartNumber: aws.Int32(part.PartNumber),
		})
	}
	out, err := b.client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(b.bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completed,
		},
	})
	if err != nil {
		return nil, mapS3Err(err)
	}
	now := time.Now().UTC()
	meta := &objectMeta{Key: key, UpdatedAt: &now}
	if out != nil && out.ETag != nil {
		meta.ETag = strings.Trim(*out.ETag, `"`)
	}
	return meta, nil
}

func (b *s3Backend) AbortMultipart(ctx context.Context, key string, uploadID string) error {
	_, err := b.client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(b.bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
	})
	return mapS3Err(err)
}

func (b *s3Backend) PresignUploadPart(ctx context.Context, key string, uploadID string, partNumber int32, ttl time.Duration) (string, map[string]string, time.Time, error) {
	if ttl <= 0 {
		ttl = time.Hour
	}
	if ttl > time.Hour {
		ttl = time.Hour
	}
	presigner := s3.NewPresignClient(b.client)
	out, err := presigner.PresignUploadPart(ctx, &s3.UploadPartInput{
		Bucket:     aws.String(b.bucket),
		Key:        aws.String(key),
		UploadId:   aws.String(uploadID),
		PartNumber: aws.Int32(partNumber),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", nil, time.Time{}, mapS3Err(err)
	}
	return out.URL, map[string]string{}, time.Now().UTC().Add(ttl), nil
}

func mapS3Err(err error) error {
	if err == nil {
		return nil
	}
	if isNotFound(err) {
		return nil
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NoSuchKey", "NotFound", "NoSuchBucket":
			return nil
		case "PreconditionFailed":
			return bizerr.NewCode(storagecap.CodeStorageObjectExists)
		}
	}
	var re *types.NotFound
	if errors.As(err, &re) {
		return nil
	}
	return err
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.ErrorCode()
		if code == "NoSuchKey" || code == "NotFound" || code == "404" || code == "NoSuchBucket" {
			return true
		}
	}
	var httpErr interface{ HTTPStatusCode() int }
	if errors.As(err, &httpErr) && httpErr.HTTPStatusCode() == http.StatusNotFound {
		return true
	}
	var nsk *types.NoSuchKey
	if errors.As(err, &nsk) {
		return true
	}
	var nf *types.NotFound
	return errors.As(err, &nf)
}

// silence unused imports in some builds
var (
	_ = bytes.MinRead
	_ = fmt.Sprintf
)
