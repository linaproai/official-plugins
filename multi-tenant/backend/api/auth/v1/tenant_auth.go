// This file declares login tenant candidate DTOs for the multi-tenant source plugin.

package v1

import "github.com/gogf/gf/v2/frame/g"

// LoginTenantsReq defines the request for login tenant candidates.
type LoginTenantsReq struct {
	g.Meta `path:"/auth/login-tenants" method:"get" tags:"Tenant Auth" summary:"Get login tenant candidates" dc:"Return tenant candidates for a user during the login tenant-selection stage." permission:"system:tenant:auth:login-tenants"`
	UserId int64 `json:"userId" v:"required" dc:"User ID" eg:"2"`
}

// LoginTenantsRes defines the login tenant candidates response.
type LoginTenantsRes struct {
	List []*LoginTenantItem `json:"list" dc:"Login tenant candidates" eg:"[]"`
}
