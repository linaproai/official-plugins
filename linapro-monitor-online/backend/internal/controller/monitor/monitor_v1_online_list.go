package monitor

import (
	"context"

	"lina-core/pkg/apitime"
	v1 "lina-plugin-linapro-monitor-online/backend/api/monitor/v1"
	monitorsvc "lina-plugin-linapro-monitor-online/backend/internal/service/monitor"
)

// OnlineList queries the online-user list through the published host session seam.
func (c *ControllerV1) OnlineList(ctx context.Context, req *v1.OnlineListReq) (res *v1.OnlineListRes, err error) {
	out, err := c.monitorSvc.List(ctx, monitorsvc.ListInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Username: req.Username,
		Ip:       req.Ip,
	})
	if err != nil {
		return nil, err
	}

	items := make([]*v1.OnlineUserItem, 0, len(out.Items))
	for _, session := range out.Items {
		items = append(items, &v1.OnlineUserItem{
			TokenId:    string(session.ID),
			Username:   session.Username,
			ClientType: session.ClientType,
			DeptName:   session.DeptName,
			Ip:         session.Ip,
			Browser:    session.Browser,
			Os:         session.Os,
			LoginTime:  apitime.Milli(session.LoginAt),
		})
	}

	return &v1.OnlineListRes{Items: items, Total: out.Total}, nil
}
