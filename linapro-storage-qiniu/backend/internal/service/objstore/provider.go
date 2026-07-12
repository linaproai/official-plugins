package objstore

import (
	"context"
	"io"
	"path"
	"strings"
	"time"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/storagecap"
	settingssvc "lina-plugin-linapro-storage-qiniu/backend/internal/service/settings"
)

// objectBackend is the cloud-specific object API used by provider.
type objectBackend interface {
	Put(ctx context.Context, key string, body io.Reader, size int64, contentType string, overwrite bool) (*objectMeta, error)
	Get(ctx context.Context, key string) (*objectMeta, io.ReadCloser, bool, error)
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string, cursor string, limit int) ([]*objectMeta, string, error)
	Stat(ctx context.Context, key string) (*objectMeta, bool, error)
	HeadBucket(ctx context.Context) error
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
