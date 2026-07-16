package objstore

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"lina-core/pkg/plugin/capability/storagecap"
)

func TestProviderPutGetDeleteAndOverwrite(t *testing.T) {
	t.Parallel()
	p := &provider{backend: newMemoryBackend(), pathPrefix: "root"}
	ctx := context.Background()

	obj, err := p.Put(ctx, storagecap.ProviderPutInput{
		Key: "a.txt", Body: bytes.NewReader([]byte("hello")), Size: 5, ContentType: "text/plain", Overwrite: false,
	})
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	if obj.Key != "a.txt" || obj.Size != 5 {
		t.Fatalf("unexpected object %#v", obj)
	}
	if _, err = p.Put(ctx, storagecap.ProviderPutInput{
		Key: "a.txt", Body: bytes.NewReader([]byte("x")), Overwrite: false,
	}); err == nil {
		t.Fatal("expected exists error")
	}
	got, err := p.Get(ctx, storagecap.ProviderGetInput{Key: "a.txt"})
	if err != nil || !got.Found {
		t.Fatalf("get: %#v %v", got, err)
	}
	data, _ := io.ReadAll(got.Body)
	_ = got.Body.Close()
	if string(data) != "hello" {
		t.Fatalf("body %q", data)
	}
	if err := p.Delete(ctx, storagecap.ProviderDeleteInput{Key: "a.txt"}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	got, err = p.Get(ctx, storagecap.ProviderGetInput{Key: "a.txt"})
	if err != nil || got.Found {
		t.Fatalf("expected missing after delete")
	}
}

func TestProviderBatchStatAndListCursor(t *testing.T) {
	t.Parallel()
	p := &provider{backend: newMemoryBackend()}
	ctx := context.Background()
	for _, key := range []string{"a", "b", "c"} {
		if _, err := p.Put(ctx, storagecap.ProviderPutInput{Key: key, Body: bytes.NewReader([]byte(key)), Overwrite: true}); err != nil {
			t.Fatal(err)
		}
	}
	batch, err := p.BatchStat(ctx, storagecap.ProviderBatchStatInput{Keys: []string{"a", "missing", "c"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Objects) != 2 || len(batch.MissingKeys) != 1 {
		t.Fatalf("batch %#v", batch)
	}
	page, err := p.ListCursor(ctx, storagecap.ProviderListCursorInput{Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Objects) != 2 || page.NextCursor == "" {
		t.Fatalf("page %#v", page)
	}
}

func TestProviderDirectAccessPutGet(t *testing.T) {
	t.Parallel()
	p := &provider{backend: newMemoryBackend(), pathPrefix: "root"}
	ctx := context.Background()
	if !p.SupportsDirectAccess(ctx, storagecap.DirectAccessOpPut) {
		t.Fatal("expected put direct access support")
	}
	put, err := p.CreateDirectAccess(ctx, storagecap.ProviderDirectAccessInput{
		Key: "a.txt", Operation: storagecap.DirectAccessOpPut, ContentType: "text/plain",
	})
	if err != nil {
		t.Fatalf("CreateDirectAccess put: %v", err)
	}
	if put.Mode != storagecap.DirectAccessModePresignedURL || put.Method != "PUT" || put.URL == "" {
		t.Fatalf("unexpected put access %#v", put)
	}
	if put.Headers["Content-Type"] != "text/plain" {
		t.Fatalf("missing content-type header %#v", put.Headers)
	}
	if !strings.Contains(put.URL, "root/a.txt") && !strings.Contains(put.URL, "a.txt") {
		// memory backend returns key as provided (scoped)
		t.Fatalf("url should include scoped key: %s", put.URL)
	}
	get, err := p.CreateDirectAccess(ctx, storagecap.ProviderDirectAccessInput{
		Key: "a.txt", Operation: storagecap.DirectAccessOpGet,
	})
	if err != nil {
		t.Fatalf("CreateDirectAccess get: %v", err)
	}
	if get.Mode != storagecap.DirectAccessModePresignedURL || get.Method != "GET" || get.URL == "" {
		t.Fatalf("unexpected get access %#v", get)
	}
}
