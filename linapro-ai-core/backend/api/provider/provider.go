// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package provider

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/provider/v1"
)

type IProviderV1 interface {
	Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error)
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
	ListProviderEndpoints(ctx context.Context, req *v1.ListProviderEndpointsReq) (res *v1.ListProviderEndpointsRes, err error)
	CreateProviderEndpoint(ctx context.Context, req *v1.CreateProviderEndpointReq) (res *v1.CreateProviderEndpointRes, err error)
	UpdateProviderEndpoint(ctx context.Context, req *v1.UpdateProviderEndpointReq) (res *v1.UpdateProviderEndpointRes, err error)
	DeleteProviderEndpoint(ctx context.Context, req *v1.DeleteProviderEndpointReq) (res *v1.DeleteProviderEndpointRes, err error)
	Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error)
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	ListModels(ctx context.Context, req *v1.ListModelsReq) (res *v1.ListModelsRes, err error)
	CreateModel(ctx context.Context, req *v1.CreateModelReq) (res *v1.CreateModelRes, err error)
	SyncModels(ctx context.Context, req *v1.SyncModelsReq) (res *v1.SyncModelsRes, err error)
	Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error)
}
