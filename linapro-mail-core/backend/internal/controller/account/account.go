// This file implements account management request handlers.

package account

import (
	"context"

	v1 "lina-plugin-linapro-mail-core/backend/api/account/v1"
	"lina-plugin-linapro-mail-core/backend/internal/model/entity"
	mailsvc "lina-plugin-linapro-mail-core/backend/internal/service/mail"
)

// List lists accounts.
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (*v1.ListRes, error) {
	out, err := c.mailSvc.ListAccounts(ctx, mailsvc.ListAccountsInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Name:     req.Name,
		Status:   req.Status,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.AccountItem, 0, len(out.List))
	for _, row := range out.List {
		items = append(items, toAccountItem(row))
	}
	return &v1.ListRes{List: items, Total: out.Total}, nil
}

// Get returns one account.
func (c *ControllerV1) Get(ctx context.Context, req *v1.GetReq) (*v1.GetRes, error) {
	row, err := c.mailSvc.GetAccount(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetRes{AccountItem: toAccountItem(row)}, nil
}

// Create creates one account.
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (*v1.CreateRes, error) {
	id, err := c.mailSvc.CreateAccount(ctx, mailsvc.CreateAccountInput{
		Name:                 req.Name,
		FromAddress:          req.FromAddress,
		OutboundConnectionID: req.OutboundConnectionId,
		InboundConnectionID:  req.InboundConnectionId,
		IsDefault:            req.IsDefault,
		Status:               req.Status,
		Remark:               req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{Id: id}, nil
}

// Update updates one account.
func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (*v1.UpdateRes, error) {
	err := c.mailSvc.UpdateAccount(ctx, mailsvc.UpdateAccountInput{
		ID:                   req.Id,
		Name:                 req.Name,
		FromAddress:          req.FromAddress,
		OutboundConnectionID: req.OutboundConnectionId,
		InboundConnectionID:  req.InboundConnectionId,
		IsDefault:            req.IsDefault,
		Status:               req.Status,
		Remark:               req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRes{}, nil
}

// Delete deletes accounts.
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (*v1.DeleteRes, error) {
	if err := c.mailSvc.DeleteAccounts(ctx, req.Ids); err != nil {
		return nil, err
	}
	return &v1.DeleteRes{}, nil
}

func toAccountItem(row *entity.Account) *v1.AccountItem {
	if row == nil {
		return nil
	}
	item := &v1.AccountItem{
		Id:                   row.Id,
		Name:                 row.Name,
		FromAddress:          row.FromAddress,
		OutboundConnectionId: row.OutboundConnectionId,
		InboundConnectionId:  row.InboundConnectionId,
		IsDefault:            row.IsDefault == mailsvc.FlagDefault,
		Status:               row.Status,
		Remark:               row.Remark,
	}
	if row.CreatedAt != nil {
		item.CreatedAt = row.CreatedAt.UnixMilli()
	}
	if row.UpdatedAt != nil {
		item.UpdatedAt = row.UpdatedAt.UnixMilli()
	}
	return item
}
