// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package model

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/model/v1"
)

type IModelV1 interface {
	ListCapabilities(ctx context.Context, req *v1.ListCapabilitiesReq) (res *v1.ListCapabilitiesRes, err error)
	UpsertCapabilities(ctx context.Context, req *v1.UpsertCapabilitiesReq) (res *v1.UpsertCapabilitiesRes, err error)
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error)
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
}
