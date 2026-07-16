// This file implements optional MultipartUploadProvider methods for the cloud
// storage provider.

package objstore

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/storagecap"
)

// SupportsMultipart reports that this cloud backend supports multipart uploads.
func (p *provider) SupportsMultipart(_ context.Context) bool {
	return p != nil && p.backend != nil
}

// CreateMultipart starts one multipart upload for a scoped key.
func (p *provider) CreateMultipart(ctx context.Context, in storagecap.ProviderMultipartCreateInput) (*storagecap.ProviderMultipartSession, error) {
	if p == nil || p.backend == nil {
		return nil, storagecap.NewMultipartUnsupportedError()
	}
	key := p.scopedKey(in.Key)
	uploadID, err := p.backend.CreateMultipart(ctx, key, in.ContentType)
	if err != nil {
		return nil, err
	}
	return &storagecap.ProviderMultipartSession{
		UploadID: uploadID,
		Key:      in.Key,
	}, nil
}

// UploadPart writes one multipart part.
func (p *provider) UploadPart(ctx context.Context, in storagecap.ProviderMultipartPartInput) (*storagecap.ProviderMultipartPartResult, error) {
	if p == nil || p.backend == nil {
		return nil, storagecap.NewMultipartUnsupportedError()
	}
	if !storagecap.ValidateMultipartPartNumber(in.PartNumber) {
		return nil, storagecap.NewMultipartPartInvalidError()
	}
	etag, err := p.backend.UploadPart(ctx, p.scopedKey(in.Key), in.UploadID, in.PartNumber, in.Body, in.Size)
	if err != nil {
		return nil, err
	}
	return &storagecap.ProviderMultipartPartResult{PartNumber: in.PartNumber, ETag: etag}, nil
}

// CompleteMultipart finishes one multipart upload.
func (p *provider) CompleteMultipart(ctx context.Context, in storagecap.ProviderMultipartCompleteInput) (*storagecap.ProviderObject, error) {
	if p == nil || p.backend == nil {
		return nil, storagecap.NewMultipartUnsupportedError()
	}
	parts := make([]completedPart, 0, len(in.Parts))
	for _, part := range in.Parts {
		parts = append(parts, completedPart{PartNumber: part.PartNumber, ETag: part.ETag})
	}
	meta, err := p.backend.CompleteMultipart(ctx, p.scopedKey(in.Key), in.UploadID, parts)
	if err != nil {
		return nil, bizerr.WrapCode(err, storagecap.CodeStorageMultipartCompleteFailed)
	}
	return p.toProviderObject(meta), nil
}

// AbortMultipart aborts one multipart upload.
func (p *provider) AbortMultipart(ctx context.Context, in storagecap.ProviderMultipartAbortInput) error {
	if p == nil || p.backend == nil {
		return storagecap.NewMultipartUnsupportedError()
	}
	return p.backend.AbortMultipart(ctx, p.scopedKey(in.Key), in.UploadID)
}

// CreateMultipartPartAccess issues a presigned URL for one multipart part.
func (p *provider) CreateMultipartPartAccess(ctx context.Context, in storagecap.ProviderMultipartPartAccessInput) (*storagecap.DirectAccess, error) {
	if p == nil || p.backend == nil {
		return nil, storagecap.NewMultipartUnsupportedError()
	}
	if !storagecap.ValidateMultipartPartNumber(in.PartNumber) || strings.TrimSpace(in.UploadID) == "" {
		return nil, storagecap.NewMultipartPartInvalidError()
	}
	url, headers, expiresAt, err := p.backend.PresignUploadPart(ctx, p.scopedKey(in.Key), in.UploadID, in.PartNumber, in.TTL)
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
}
