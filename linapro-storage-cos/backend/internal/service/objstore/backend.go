package objstore

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/storagecap"
	settingssvc "lina-plugin-linapro-storage-cos/backend/internal/service/settings"
)

type cosBackend struct {
	client *cos.Client
	bucket string
}

func newCloudBackend(snapshot *settingssvc.Snapshot) (objectBackend, error) {
	if snapshot == nil {
		return nil, bizerr.NewCode(settingssvc.CodeConfigInvalid)
	}
	region := strings.TrimSpace(snapshot.Region)
	bucket := strings.TrimSpace(snapshot.Bucket)
	if region == "" || bucket == "" {
		return nil, bizerr.NewCode(settingssvc.CodeConfigInvalid)
	}
	endpoint := strings.TrimSpace(snapshot.Endpoint)
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucket, region)
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, bizerr.WrapCode(err, settingssvc.CodeConfigInvalid)
	}
	baseURL := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(baseURL, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  strings.TrimSpace(snapshot.AccessKeyID),
			SecretKey: strings.TrimSpace(snapshot.SecretAccessKey),
		},
	})
	return &cosBackend{client: client, bucket: bucket}, nil
}

// readerWithoutClose exposes only io.Reader so SDKs cannot Close host-owned streams.
type readerWithoutClose struct {
	io.Reader
}

// readSeekerWithoutClose exposes Read+Seek without Close for SDK retries.
type readSeekerWithoutClose struct {
	io.ReadSeeker
}

// asStorageBodyWithoutClose wraps body so cloud SDKs cannot Close a caller-owned stream.
func asStorageBodyWithoutClose(body io.Reader) io.Reader {
	if body == nil {
		return nil
	}
	if seeker, ok := body.(io.ReadSeeker); ok {
		return readSeekerWithoutClose{ReadSeeker: seeker}
	}
	return readerWithoutClose{Reader: body}
}

func (b *cosBackend) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string, overwrite bool) (*objectMeta, error) {
	if !overwrite {
		_, found, err := b.Stat(ctx, key)
		if err != nil {
			return nil, err
		}
		if found {
			return nil, bizerr.NewCode(storagecap.CodeStorageObjectExists)
		}
	}
	// Always set ContentLength from the host-provided size. After the host wraps
	// the body as a non-*os.File ReadSeeker, the COS SDK cannot infer length via
	// Stat and would otherwise fall back to chunked upload or fail length checks.
	header := &cos.ObjectPutHeaderOptions{}
	if size >= 0 {
		header.ContentLength = size
	}
	if strings.TrimSpace(contentType) != "" {
		header.ContentType = contentType
	}
	opt := &cos.ObjectPutOptions{ObjectPutHeaderOptions: header}
	// Strip Close so the SDK TeeReader cannot close a caller-owned ReadCloser.
	resp, err := b.client.Object.Put(ctx, key, asStorageBodyWithoutClose(body), opt)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	etag := ""
	if resp != nil && resp.Header != nil {
		etag = strings.Trim(resp.Header.Get("ETag"), `"`)
	}
	return &objectMeta{Key: key, Size: size, ContentType: contentType, ETag: etag, UpdatedAt: &now}, nil
}

func (b *cosBackend) Get(ctx context.Context, key string) (*objectMeta, io.ReadCloser, bool, error) {
	resp, err := b.client.Object.Get(ctx, key, nil)
	if err != nil {
		if cos.IsNotFoundError(err) {
			return nil, nil, false, nil
		}
		return nil, nil, false, err
	}
	meta := &objectMeta{Key: key, Size: resp.ContentLength, ContentType: resp.Header.Get("Content-Type"), ETag: strings.Trim(resp.Header.Get("ETag"), `"`)}
	if lm := resp.Header.Get("Last-Modified"); lm != "" {
		if t, err := http.ParseTime(lm); err == nil {
			t = t.UTC()
			meta.UpdatedAt = &t
		}
	}
	return meta, resp.Body, true, nil
}

func (b *cosBackend) Delete(ctx context.Context, key string) error {
	_, err := b.client.Object.Delete(ctx, key)
	if err != nil && cos.IsNotFoundError(err) {
		return nil
	}
	return err
}

func (b *cosBackend) List(ctx context.Context, prefix string, cursor string, limit int) ([]*objectMeta, string, error) {
	opt := &cos.BucketGetOptions{Prefix: prefix, MaxKeys: limit}
	if cursor != "" {
		opt.Marker = cursor
	}
	result, _, err := b.client.Bucket.Get(ctx, opt)
	if err != nil {
		return nil, "", err
	}
	metas := make([]*objectMeta, 0, len(result.Contents))
	for _, item := range result.Contents {
		meta := &objectMeta{Key: item.Key, Size: item.Size, ETag: strings.Trim(item.ETag, `"`)}
		if item.LastModified != "" {
			if t, err := time.Parse(time.RFC3339, item.LastModified); err == nil {
				t = t.UTC()
				meta.UpdatedAt = &t
			}
		}
		metas = append(metas, meta)
	}
	next := ""
	if result.IsTruncated && len(metas) > 0 {
		next = metas[len(metas)-1].Key
	}
	return metas, next, nil
}

func (b *cosBackend) Stat(ctx context.Context, key string) (*objectMeta, bool, error) {
	resp, err := b.client.Object.Head(ctx, key, nil)
	if err != nil {
		if cos.IsNotFoundError(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	meta := &objectMeta{Key: key, Size: resp.ContentLength, ContentType: resp.Header.Get("Content-Type"), ETag: strings.Trim(resp.Header.Get("ETag"), `"`)}
	if lm := resp.Header.Get("Last-Modified"); lm != "" {
		if t, err := http.ParseTime(lm); err == nil {
			t = t.UTC()
			meta.UpdatedAt = &t
		}
	}
	return meta, true, nil
}

func (b *cosBackend) HeadBucket(ctx context.Context) error {
	_, err := b.client.Bucket.Head(ctx)
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

func (b *cosBackend) PresignPut(ctx context.Context, key string, contentType string, ttl time.Duration) (string, map[string]string, time.Time, error) {
	ttl = normalizePresignTTL(ttl)
	var opt *cos.PresignedURLOptions
	headers := map[string]string{}
	if strings.TrimSpace(contentType) != "" {
		h := http.Header{}
		h.Set("Content-Type", contentType)
		opt = &cos.PresignedURLOptions{Header: &h}
		headers["Content-Type"] = contentType
	}
	u, err := b.client.Object.GetPresignedURL2(ctx, http.MethodPut, key, ttl, opt)
	if err != nil {
		return "", nil, time.Time{}, err
	}
	return u.String(), headers, time.Now().UTC().Add(ttl), nil
}

func (b *cosBackend) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, time.Time, error) {
	ttl = normalizePresignTTL(ttl)
	u, err := b.client.Object.GetPresignedURL2(ctx, http.MethodGet, key, ttl, nil)
	if err != nil {
		return "", time.Time{}, err
	}
	return u.String(), time.Now().UTC().Add(ttl), nil
}
