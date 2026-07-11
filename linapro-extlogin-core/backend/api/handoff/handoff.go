// Package handoff defines the public handoff exchange controller contract.
package handoff

import (
	"context"

	v1 "lina-plugin-linapro-extlogin-core/backend/api/handoff/v1"
)

// IHandoffV1 is the public handoff exchange API.
type IHandoffV1 interface {
	ExchangeLoginHandoff(ctx context.Context, req *v1.ExchangeLoginHandoffReq) (res *v1.ExchangeLoginHandoffRes, err error)
}
