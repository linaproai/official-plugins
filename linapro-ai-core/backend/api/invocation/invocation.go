// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package invocation

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/invocation/v1"
)

type IInvocationV1 interface {
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
}
