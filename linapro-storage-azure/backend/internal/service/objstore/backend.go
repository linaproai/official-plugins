// Cloud object backend for Azure Blob Storage using the official Azure SDK.
//
// This file maps storagecap object operations onto Azure Blob container APIs
// (account + container model) and normalizes not-found responses.
package objstore

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/storagecap"
	settingssvc "lina-plugin-linapro-storage-azure/backend/internal/service/settings"
)

// azureBackend implements objectBackend against one Azure Blob container.
type azureBackend struct {
	client    *azblob.Client
	container string
}

// newCloudBackend constructs an Azure Blob client from settings.
func newCloudBackend(snapshot *settingssvc.Snapshot) (objectBackend, error) {
	if snapshot == nil {
		return nil, bizerr.NewCode(settingssvc.CodeConfigInvalid)
	}
	accountName := strings.TrimSpace(snapshot.AccountName)
	accountKey := strings.TrimSpace(snapshot.AccountKey)
	container := strings.TrimSpace(snapshot.Container)
	if accountName == "" || accountKey == "" || container == "" {
		return nil, bizerr.NewCode(settingssvc.CodeConfigInvalid)
	}
	serviceURL := strings.TrimSpace(snapshot.Endpoint)
	if serviceURL == "" {
		serviceURL = "https://" + accountName + ".blob.core.windows.net/"
	}
	if !strings.HasSuffix(serviceURL, "/") {
		serviceURL += "/"
	}
	cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return nil, bizerr.WrapCode(err, settingssvc.CodeConfigInvalid)
	}
	client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	if err != nil {
		return nil, bizerr.WrapCode(err, settingssvc.CodeConfigInvalid)
	}
	return &azureBackend{client: client, container: container}, nil
}

func (b *azureBackend) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string, overwrite bool) (*objectMeta, error) {
	if !overwrite {
		_, found, err := b.Stat(ctx, key)
		if err != nil {
			return nil, err
		}
		if found {
			return nil, bizerr.NewCode(storagecap.CodeStorageObjectExists)
		}
	}
	var opts *azblob.UploadStreamOptions
	if strings.TrimSpace(contentType) != "" {
		opts = &azblob.UploadStreamOptions{
			HTTPHeaders: &blob.HTTPHeaders{
				BlobContentType: to.Ptr(contentType),
			},
		}
	}
	_, err := b.client.UploadStream(ctx, b.container, key, body, opts)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	return &objectMeta{Key: key, Size: size, ContentType: contentType, UpdatedAt: &now}, nil
}

func (b *azureBackend) Get(ctx context.Context, key string) (*objectMeta, io.ReadCloser, bool, error) {
	resp, err := b.client.DownloadStream(ctx, b.container, key, nil)
	if err != nil {
		if isAzureNotFound(err) {
			return nil, nil, false, nil
		}
		return nil, nil, false, err
	}
	meta := &objectMeta{Key: key}
	if resp.ContentLength != nil {
		meta.Size = *resp.ContentLength
	}
	if resp.ContentType != nil {
		meta.ContentType = *resp.ContentType
	}
	if resp.ETag != nil {
		meta.ETag = strings.Trim(string(*resp.ETag), `"`)
	}
	if resp.LastModified != nil {
		t := resp.LastModified.UTC()
		meta.UpdatedAt = &t
	}
	return meta, resp.Body, true, nil
}

func (b *azureBackend) Delete(ctx context.Context, key string) error {
	_, err := b.client.DeleteBlob(ctx, b.container, key, nil)
	if err != nil && isAzureNotFound(err) {
		return nil
	}
	return err
}

func (b *azureBackend) List(ctx context.Context, prefix string, cursor string, limit int) ([]*objectMeta, string, error) {
	maxResults := int32(limit)
	if maxResults <= 0 {
		maxResults = int32(storagecap.DefaultListLimit)
	}
	opts := &azblob.ListBlobsFlatOptions{
		MaxResults: &maxResults,
	}
	if prefix != "" {
		opts.Prefix = &prefix
	}
	if cursor != "" {
		opts.Marker = &cursor
	}
	pager := b.client.NewListBlobsFlatPager(b.container, opts)
	if !pager.More() {
		return []*objectMeta{}, "", nil
	}
	page, err := pager.NextPage(ctx)
	if err != nil {
		return nil, "", err
	}
	metas := make([]*objectMeta, 0, len(page.Segment.BlobItems))
	for _, item := range page.Segment.BlobItems {
		if item == nil || item.Name == nil {
			continue
		}
		meta := &objectMeta{Key: *item.Name}
		if item.Properties != nil {
			if item.Properties.ContentLength != nil {
				meta.Size = *item.Properties.ContentLength
			}
			if item.Properties.ContentType != nil {
				meta.ContentType = *item.Properties.ContentType
			}
			if item.Properties.ETag != nil {
				meta.ETag = strings.Trim(string(*item.Properties.ETag), `"`)
			}
			if item.Properties.LastModified != nil {
				t := item.Properties.LastModified.UTC()
				meta.UpdatedAt = &t
			}
		}
		metas = append(metas, meta)
	}
	next := ""
	if page.NextMarker != nil && strings.TrimSpace(*page.NextMarker) != "" {
		next = *page.NextMarker
	}
	return metas, next, nil
}

func (b *azureBackend) Stat(ctx context.Context, key string) (*objectMeta, bool, error) {
	blobClient := b.client.ServiceClient().NewContainerClient(b.container).NewBlobClient(key)
	props, err := blobClient.GetProperties(ctx, nil)
	if err != nil {
		if isAzureNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	meta := &objectMeta{Key: key}
	if props.ContentLength != nil {
		meta.Size = *props.ContentLength
	}
	if props.ContentType != nil {
		meta.ContentType = *props.ContentType
	}
	if props.ETag != nil {
		meta.ETag = strings.Trim(string(*props.ETag), `"`)
	}
	if props.LastModified != nil {
		t := props.LastModified.UTC()
		meta.UpdatedAt = &t
	}
	return meta, true, nil
}

func (b *azureBackend) HeadBucket(ctx context.Context) error {
	_, err := b.client.ServiceClient().NewContainerClient(b.container).GetProperties(ctx, nil)
	return err
}

func isAzureNotFound(err error) bool {
	if err == nil {
		return false
	}
	if bloberror.HasCode(err, bloberror.BlobNotFound, bloberror.ContainerNotFound) {
		return true
	}
	var respErr *azcore.ResponseError
	if errors.As(err, &respErr) {
		return respErr.StatusCode == http.StatusNotFound
	}
	return false
}
