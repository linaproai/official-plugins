package objstore

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/storagecap"
)

// memoryBackend is used by unit tests.
type memoryBackend struct {
	objects    map[string]*memoryObject
	multiparts map[string]*memoryMultipart
}

type memoryObject struct {
	body        []byte
	contentType string
	etag        string
	updatedAt   time.Time
}

type memoryMultipart struct {
	key   string
	parts map[int32][]byte
}

func newMemoryBackend() *memoryBackend {
	return &memoryBackend{
		objects:    map[string]*memoryObject{},
		multiparts: map[string]*memoryMultipart{},
	}
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

func (m *memoryBackend) CreateMultipart(_ context.Context, key string, _ string) (string, error) {
	if m.multiparts == nil {
		m.multiparts = make(map[string]*memoryMultipart)
	}
	uploadID := fmt.Sprintf("upload-%d", len(m.multiparts)+1)
	m.multiparts[uploadID] = &memoryMultipart{key: key, parts: make(map[int32][]byte)}
	return uploadID, nil
}

func (m *memoryBackend) UploadPart(_ context.Context, _ string, uploadID string, partNumber int32, body io.Reader, _ int64) (string, error) {
	mp := m.multiparts[uploadID]
	if mp == nil {
		return "", fmt.Errorf("multipart not found")
	}
	data, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}
	mp.parts[partNumber] = data
	return fmt.Sprintf("etag-%d", partNumber), nil
}

func (m *memoryBackend) CompleteMultipart(_ context.Context, key string, uploadID string, parts []completedPart) (*objectMeta, error) {
	mp := m.multiparts[uploadID]
	if mp == nil {
		return nil, fmt.Errorf("multipart not found")
	}
	var body []byte
	for _, part := range parts {
		body = append(body, mp.parts[part.PartNumber]...)
	}
	now := time.Now().UTC()
	if m.objects == nil {
		m.objects = make(map[string]*memoryObject)
	}
	m.objects[key] = &memoryObject{body: body, etag: "mp-etag", updatedAt: now}
	delete(m.multiparts, uploadID)
	return &objectMeta{Key: key, Size: int64(len(body)), ETag: "mp-etag", UpdatedAt: &now}, nil
}

func (m *memoryBackend) AbortMultipart(_ context.Context, _ string, uploadID string) error {
	delete(m.multiparts, uploadID)
	return nil
}

func (m *memoryBackend) PresignUploadPart(_ context.Context, key string, uploadID string, partNumber int32, ttl time.Duration) (string, map[string]string, time.Time, error) {
	if ttl <= 0 {
		ttl = time.Hour
	}
	return fmt.Sprintf("https://memory.test/part/%s/%s/%d", key, uploadID, partNumber), map[string]string{}, time.Now().UTC().Add(ttl), nil
}
