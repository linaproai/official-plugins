// demo_v1_record_create.go implements the linapro-demo-source record create HTTP handler.

package demo

import (
	"context"

	"lina-plugin-linapro-demo-source/backend/api/demo/v1"

	"github.com/gogf/gf/v2/frame/g"

	demosvc "lina-plugin-linapro-demo-source/backend/internal/service/demo"
)

// CreateRecord creates one demo record with an optional attachment.
func (c *ControllerV1) CreateRecord(
	ctx context.Context,
	req *v1.CreateRecordReq,
) (res *v1.CreateRecordRes, err error) {
	out, err := c.demoSvc.CreateRecord(ctx, &demosvc.CreateRecordInput{
		Title:   req.Title,
		Content: req.Content,
		File:    g.RequestFromCtx(ctx).GetUploadFile("file"),
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateRecordRes{Id: out.Id}, nil
}
