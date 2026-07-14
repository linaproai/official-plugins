// This file implements connection management request handlers.

package connection

import (
	"context"

	v1 "lina-plugin-linapro-mail-core/backend/api/connection/v1"
	"lina-plugin-linapro-mail-core/backend/internal/model/entity"
	mailsvc "lina-plugin-linapro-mail-core/backend/internal/service/mail"
)

// List lists connections.
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (*v1.ListRes, error) {
	out, err := c.mailSvc.ListConnections(ctx, mailsvc.ListConnectionsInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Name:     req.Name,
		Kind:     req.Kind,
		Status:   req.Status,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.ConnectionItem, 0, len(out.List))
	for _, row := range out.List {
		items = append(items, toConnectionItem(row))
	}
	return &v1.ListRes{List: items, Total: out.Total}, nil
}

// Get returns one connection.
func (c *ControllerV1) Get(ctx context.Context, req *v1.GetReq) (*v1.GetRes, error) {
	row, err := c.mailSvc.GetConnection(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetRes{ConnectionItem: toConnectionItem(row)}, nil
}

// Create creates one connection.
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (*v1.CreateRes, error) {
	id, err := c.mailSvc.CreateConnection(ctx, mailsvc.CreateConnectionInput{
		Name:      req.Name,
		Kind:      req.Kind,
		Host:      req.Host,
		Port:      req.Port,
		Username:  req.Username,
		SecretRef: req.SecretRef,
		TLSMode:   req.TlsMode,
		AuthMode:  req.AuthMode,
		ExtraJSON: req.ExtraJson,
		Status:    req.Status,
		Remark:    req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{Id: id}, nil
}

// Update updates one connection.
func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (*v1.UpdateRes, error) {
	err := c.mailSvc.UpdateConnection(ctx, mailsvc.UpdateConnectionInput{
		ID:        req.Id,
		Name:      req.Name,
		Kind:      req.Kind,
		Host:      req.Host,
		Port:      req.Port,
		Username:  req.Username,
		SecretRef: req.SecretRef,
		TLSMode:   req.TlsMode,
		AuthMode:  req.AuthMode,
		ExtraJSON: req.ExtraJson,
		Status:    req.Status,
		Remark:    req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRes{}, nil
}

// Delete deletes connections.
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (*v1.DeleteRes, error) {
	if err := c.mailSvc.DeleteConnections(ctx, req.Ids); err != nil {
		return nil, err
	}
	return &v1.DeleteRes{}, nil
}

// Probe probes one connection.
func (c *ControllerV1) Probe(ctx context.Context, req *v1.ProbeReq) (*v1.ProbeRes, error) {
	if err := c.mailSvc.ProbeConnection(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.ProbeRes{}, nil
}

func toConnectionItem(row *entity.Connection) *v1.ConnectionItem {
	if row == nil {
		return nil
	}
	item := &v1.ConnectionItem{
		Id:        row.Id,
		Name:      row.Name,
		Kind:      row.Kind,
		Host:      row.Host,
		Port:      row.Port,
		Username:  row.Username,
		SecretRef: row.SecretRef,
		TlsMode:   row.TlsMode,
		AuthMode:  row.AuthMode,
		ExtraJson: row.ExtraJson,
		Status:    row.Status,
		Remark:    row.Remark,
	}
	if row.CreatedAt != nil {
		item.CreatedAt = row.CreatedAt.UnixMilli()
	}
	if row.UpdatedAt != nil {
		item.UpdatedAt = row.UpdatedAt.UnixMilli()
	}
	return item
}
