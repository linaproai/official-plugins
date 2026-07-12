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
