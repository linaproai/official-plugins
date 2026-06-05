// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package invocation

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/invocation/v1"
)

type IInvocationV1 interface {
	Clean(ctx context.Context, req *v1.CleanReq) (res *v1.CleanRes, err error)
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	ListProviderOperations(ctx context.Context, req *v1.ListProviderOperationsReq) (res *v1.ListProviderOperationsRes, err error)
	GetProviderOperation(ctx context.Context, req *v1.GetProviderOperationReq) (res *v1.GetProviderOperationRes, err error)
}
