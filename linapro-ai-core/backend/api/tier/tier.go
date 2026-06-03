// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package tier

import (
	"context"

	"lina-plugin-linapro-ai-core/backend/api/tier/v1"
)

type ITierV1 interface {
	ListMethodDefaults(ctx context.Context, req *v1.ListMethodDefaultsReq) (res *v1.ListMethodDefaultsRes, err error)
	UpdateMethodDefault(ctx context.Context, req *v1.UpdateMethodDefaultReq) (res *v1.UpdateMethodDefaultRes, err error)
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	Test(ctx context.Context, req *v1.TestReq) (res *v1.TestRes, err error)
	Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error)
}
