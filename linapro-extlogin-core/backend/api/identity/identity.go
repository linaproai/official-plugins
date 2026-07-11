// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package identity

import (
	"context"

	"lina-plugin-linapro-extlogin-core/backend/api/identity/v1"
)

type IIdentityV1 interface {
	BindIdentity(ctx context.Context, req *v1.BindIdentityReq) (res *v1.BindIdentityRes, err error)
	ListIdentities(ctx context.Context, req *v1.ListIdentitiesReq) (res *v1.ListIdentitiesRes, err error)
	UnbindIdentity(ctx context.Context, req *v1.UnbindIdentityReq) (res *v1.UnbindIdentityRes, err error)
}
