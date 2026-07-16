package objstore

import (
	"context"
	"io"
	"path"
	"strings"
	"time"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/storagecap"
	settingssvc "lina-plugin-linapro-storage-cos/backend/internal/service/settings"
)

// objectBackend is the cloud-specific object API used by provider.
type objectBackend interface {
	// Put writes one object key and returns provider metadata.
	Put(ctx context.Context, key string, body io.Reader, size int64, contentType string, overwrite bool) (*objectMeta, error)
	// Get reads one object key. A missing object returns found=false.
	Get(ctx context.Context, key string) (*objectMeta, io.ReadCloser, bool, error)
	// Delete removes one object key. Deleting a missing object is a no-op.
	Delete(ctx context.Context, key string) error
	// List lists object keys under one bounded prefix with optional cursor.
	List(ctx context.Context, prefix string, cursor string, limit int) ([]*objectMeta, string, error)
	// Stat reads one object key metadata. A missing object returns found=false.
	Stat(ctx context.Context, key string) (*objectMeta, bool, error)
	// HeadBucket probes bucket connectivity without mutating objects.
	HeadBucket(ctx context.Context) error
	// PresignPut issues time-limited client upload access (presigned PUT URL or form-post fields).
	PresignPut(ctx context.Context, key string, contentType string, ttl time.Duration) (url string, headers map[string]string, expiresAt time.Time, err error)
	// PresignGet returns a time-limited GET URL for client direct download.
	PresignGet(ctx context.Context, key string, ttl time.Duration) (url string, expiresAt time.Time, err error)
}

type objectMeta struct {
	Key         string
	Size        int64
	ContentType string
	ETag        string
	UpdatedAt   *time.Time
}

type provider struct {
	backend    objectBackend
	pathPrefix string
}

// NewProvider constructs a storagecap.Provider from settings.
func NewProvider(snapshot *settingssvc.Snapshot) (storagecap.Provider, error) {
	if snapshot == nil {
		return nil, bizerr.NewCode(settingssvc.CodeConfigInvalid)
	}
	backend, err := newCloudBackend(snapshot)
	if err != nil {
		return nil, err
	}
	return &provider{
		backend:    backend,
		pathPrefix: settingssvc.NormalizePathPrefix(snapshot.PathPrefix),
	}, nil
}

// Probe verifies bucket connectivity.
func Probe(ctx context.Context, snapshot *settingssvc.Snapshot) error {
	backend, err := newCloudBackend(snapshot)
	if err != nil {
		return err
	}
	if err := backend.HeadBucket(ctx); err != nil {
		return bizerr.WrapCode(err, settingssvc.CodeTestFailed)
	}
	return nil
}

func (p *provider) scopedKey(key string) string {
	key = strings.Trim(strings.ReplaceAll(strings.TrimSpace(key), "\\", "/"), "/")
	if p.pathPrefix == "" {
		return key
	}
	if key == "" {
		return p.pathPrefix
	}
	return path.Join(p.pathPrefix, key)
}

func (p *provider) unscopedKey(key string) string {
	key = strings.Trim(strings.ReplaceAll(strings.TrimSpace(key), "\\", "/"), "/")
	if p.pathPrefix == "" {
		return key
	}
	prefix := p.pathPrefix + "/"
	if key == p.pathPrefix {
		return ""
	}
	if strings.HasPrefix(key, prefix) {
		return strings.TrimPrefix(key, prefix)
	}
	return key
}

func (p *provider) toProviderObject(meta *objectMeta) *storagecap.ProviderObject {
	if meta == nil {
		return nil
	}
	return &storagecap.ProviderObject{
		Key:         p.unscopedKey(meta.Key),
		Size:        meta.Size,
		ContentType: meta.ContentType,
		ETag:        meta.ETag,
		UpdatedAt:   meta.UpdatedAt,
		Visibility:  storagecap.VisibilityPrivate,
	}
}

func (p *provider) Put(ctx context.Context, in storagecap.ProviderPutInput) (*storagecap.ProviderObject, error) {
	meta, err := p.backend.Put(ctx, p.scopedKey(in.Key), in.Body, in.Size, in.ContentType, in.Overwrite)
	if err != nil {
		return nil, err
	}
	return p.toProviderObject(meta), nil
}

func (p *provider) Get(ctx context.Context, in storagecap.ProviderGetInput) (*storagecap.ProviderGetOutput, error) {
	meta, body, found, err := p.backend.Get(ctx, p.scopedKey(in.Key))
	if err != nil {
		return nil, err
	}
	if !found {
		return &storagecap.ProviderGetOutput{Found: false}, nil
	}
	return &storagecap.ProviderGetOutput{Object: p.toProviderObject(meta), Body: body, Found: true}, nil
}

func (p *provider) Delete(ctx context.Context, in storagecap.ProviderDeleteInput) error {
	return p.backend.Delete(ctx, p.scopedKey(in.Key))
}

func (p *provider) DeleteMany(ctx context.Context, in storagecap.ProviderDeleteManyInput) error {
	for _, key := range in.Keys {
		if err := p.backend.Delete(ctx, p.scopedKey(key)); err != nil {
			return err
		}
	}
	return nil
}

func (p *provider) List(ctx context.Context, in storagecap.ProviderListInput) (*storagecap.ProviderListOutput, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = storagecap.DefaultListLimit
	}
	if limit > storagecap.MaxListLimit {
		limit = storagecap.MaxListLimit
	}
	metas, _, err := p.backend.List(ctx, p.scopedKey(in.Prefix), "", limit)
	if err != nil {
		return nil, err
	}
	objects := make([]*storagecap.ProviderObject, 0, len(metas))
	for _, meta := range metas {
		objects = append(objects, p.toProviderObject(meta))
	}
	return &storagecap.ProviderListOutput{Objects: objects}, nil
}

func (p *provider) ListCursor(ctx context.Context, in storagecap.ProviderListCursorInput) (*storagecap.ProviderListCursorOutput, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = storagecap.DefaultListLimit
	}
	if limit > storagecap.MaxListLimit {
		limit = storagecap.MaxListLimit
	}
	cursor := ""
	if strings.TrimSpace(in.Cursor) != "" {
		cursor = p.scopedKey(in.Cursor)
	}
	metas, next, err := p.backend.List(ctx, p.scopedKey(in.Prefix), cursor, limit)
	if err != nil {
		return nil, err
	}
	objects := make([]*storagecap.ProviderObject, 0, len(metas))
	for _, meta := range metas {
		objects = append(objects, p.toProviderObject(meta))
	}
	nextCursor := ""
	if next != "" {
		nextCursor = p.unscopedKey(next)
	}
	return &storagecap.ProviderListCursorOutput{Objects: objects, NextCursor: nextCursor}, nil
}

func (p *provider) Stat(ctx context.Context, in storagecap.ProviderStatInput) (*storagecap.ProviderStatOutput, error) {
	meta, found, err := p.backend.Stat(ctx, p.scopedKey(in.Key))
	if err != nil {
		return nil, err
	}
	if !found {
		return &storagecap.ProviderStatOutput{Found: false}, nil
	}
	return &storagecap.ProviderStatOutput{Object: p.toProviderObject(meta), Found: true}, nil
}

func (p *provider) BatchStat(ctx context.Context, in storagecap.ProviderBatchStatInput) (*storagecap.ProviderBatchStatOutput, error) {
	out := &storagecap.ProviderBatchStatOutput{Objects: []*storagecap.ProviderObject{}}
	for _, key := range in.Keys {
		meta, found, err := p.backend.Stat(ctx, p.scopedKey(key))
		if err != nil {
			return nil, err
		}
		if !found {
			out.MissingKeys = append(out.MissingKeys, key)
			continue
		}
		out.Objects = append(out.Objects, p.toProviderObject(meta))
	}
	return out, nil
}

// SupportsDirectAccess reports that this cloud backend can issue client direct access.
func (p *provider) SupportsDirectAccess(_ context.Context, op storagecap.DirectAccessOperation) bool {
	switch storagecap.NormalizeDirectAccessOperation(op) {
	case storagecap.DirectAccessOpPut, storagecap.DirectAccessOpGet:
		return p != nil && p.backend != nil
	default:
		return false
	}
}

// CreateDirectAccess issues presigned put/get URLs for the scoped object key.
func (p *provider) CreateDirectAccess(ctx context.Context, in storagecap.ProviderDirectAccessInput) (*storagecap.DirectAccess, error) {
	if p == nil || p.backend == nil {
		return nil, bizerr.NewCode(storagecap.CodeStorageDirectAccessIssueFailed)
	}
	op := storagecap.NormalizeDirectAccessOperation(in.Operation)
	key := p.scopedKey(in.Key)
	switch op {
	case storagecap.DirectAccessOpPut:
		url, headers, expiresAt, err := p.backend.PresignPut(ctx, key, in.ContentType, in.TTL)
		if err != nil {
			return nil, bizerr.WrapCode(err, storagecap.CodeStorageDirectAccessIssueFailed)
		}
		return &storagecap.DirectAccess{
			Mode:      storagecap.DirectAccessModePresignedURL,
			Operation: storagecap.DirectAccessOpPut,
			Method:    "PUT",
			URL:       url,
			Headers:   headers,
			ExpiresAt: expiresAt,
		}, nil
	case storagecap.DirectAccessOpGet:
		url, expiresAt, err := p.backend.PresignGet(ctx, key, in.TTL)
		if err != nil {
			return nil, bizerr.WrapCode(err, storagecap.CodeStorageDirectAccessIssueFailed)
		}
		return &storagecap.DirectAccess{
			Mode:      storagecap.DirectAccessModePresignedURL,
			Operation: storagecap.DirectAccessOpGet,
			Method:    "GET",
			URL:       url,
			ExpiresAt: expiresAt,
		}, nil
	default:
		return nil, bizerr.NewCode(storagecap.CodeStorageDirectAccessOperationInvalid)
	}
}
