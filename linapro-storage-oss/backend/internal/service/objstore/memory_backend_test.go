package objstore

import (
	"bytes"
	"context"
	"io"
	"sort"
	"strings"
	"time"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/storagecap"
)

// memoryBackend is used by unit tests.
type memoryBackend struct {
	objects map[string]*memoryObject
}

type memoryObject struct {
	body        []byte
	contentType string
	etag        string
	updatedAt   time.Time
}

func newMemoryBackend() *memoryBackend {
	return &memoryBackend{objects: map[string]*memoryObject{}}
}

func (m *memoryBackend) Put(_ context.Context, key string, body io.Reader, _ int64, contentType string, overwrite bool) (*objectMeta, error) {
	if !overwrite {
		if _, ok := m.objects[key]; ok {
			return nil, bizerr.NewCode(storagecap.CodeStorageObjectExists)
		}
	}
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	if data == nil {
		data = []byte{}
	}
	now := time.Now().UTC()
	m.objects[key] = &memoryObject{body: data, contentType: contentType, etag: "mem", updatedAt: now}
	return &objectMeta{Key: key, Size: int64(len(data)), ContentType: contentType, ETag: "mem", UpdatedAt: &now}, nil
}

func (m *memoryBackend) Get(_ context.Context, key string) (*objectMeta, io.ReadCloser, bool, error) {
	obj, ok := m.objects[key]
	if !ok {
		return nil, nil, false, nil
	}
	updated := obj.updatedAt
	return &objectMeta{Key: key, Size: int64(len(obj.body)), ContentType: obj.contentType, ETag: obj.etag, UpdatedAt: &updated},
		io.NopCloser(bytes.NewReader(obj.body)), true, nil
}

func (m *memoryBackend) Delete(_ context.Context, key string) error {
	delete(m.objects, key)
	return nil
}

func (m *memoryBackend) List(_ context.Context, prefix string, cursor string, limit int) ([]*objectMeta, string, error) {
	keys := make([]string, 0, len(m.objects))
	for key := range m.objects {
		if prefix == "" || strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	if cursor != "" {
		filtered := keys[:0]
		for _, key := range keys {
			if key > cursor {
				filtered = append(filtered, key)
			}
		}
		keys = filtered
	}
	var next string
	if limit > 0 && len(keys) > limit {
		next = keys[limit-1]
		keys = keys[:limit]
	}
	out := make([]*objectMeta, 0, len(keys))
	for _, key := range keys {
		obj := m.objects[key]
		updated := obj.updatedAt
		out = append(out, &objectMeta{Key: key, Size: int64(len(obj.body)), ContentType: obj.contentType, ETag: obj.etag, UpdatedAt: &updated})
	}
	return out, next, nil
}

func (m *memoryBackend) Stat(_ context.Context, key string) (*objectMeta, bool, error) {
	obj, ok := m.objects[key]
	if !ok {
		return nil, false, nil
	}
	updated := obj.updatedAt
	return &objectMeta{Key: key, Size: int64(len(obj.body)), ContentType: obj.contentType, ETag: obj.etag, UpdatedAt: &updated}, true, nil
}

func (m *memoryBackend) HeadBucket(context.Context) error { return nil }

func (m *memoryBackend) PresignPut(_ context.Context, key string, contentType string, ttl time.Duration) (string, map[string]string, time.Time, error) {
	if ttl <= 0 {
		ttl = time.Hour
	}
	headers := map[string]string{}
	if strings.TrimSpace(contentType) != "" {
		headers["Content-Type"] = contentType
	}
	return "https://memory.test/put/" + key, headers, time.Now().UTC().Add(ttl), nil
}

func (m *memoryBackend) PresignGet(_ context.Context, key string, ttl time.Duration) (string, time.Time, error) {
	if ttl <= 0 {
		ttl = time.Hour
	}
	return "https://memory.test/get/" + key, time.Now().UTC().Add(ttl), nil
}
