// demo_v1_record_update.go implements the linapro-demo-source record update HTTP handler.

package demo

import (
	"context"

	"lina-plugin-linapro-demo-source/backend/api/demo/v1"

	"github.com/gogf/gf/v2/frame/g"

	demosvc "lina-plugin-linapro-demo-source/backend/internal/service/demo"
)

// UpdateRecord updates one demo record and optionally replaces its attachment.
func (c *ControllerV1) UpdateRecord(
	ctx context.Context,
	req *v1.UpdateRecordReq,
) (res *v1.UpdateRecordRes, err error) {
	out, err := c.demoSvc.UpdateRecord(ctx, &demosvc.UpdateRecordInput{
		Id:               req.Id,
		Title:            req.Title,
		Content:          req.Content,
		File:             g.RequestFromCtx(ctx).GetUploadFile("file"),
		RemoveAttachment: req.RemoveAttachment == 1,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRecordRes{Id: out.Id}, nil
}
