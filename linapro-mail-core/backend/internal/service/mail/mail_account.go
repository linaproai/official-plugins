// This file implements Account CRUD and default account resolution.

package mail

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-plugin-linapro-mail-core/backend/cap/mailcap"
	"lina-plugin-linapro-mail-core/backend/internal/dao"
	"lina-plugin-linapro-mail-core/backend/internal/model/do"
	"lina-plugin-linapro-mail-core/backend/internal/model/entity"
)

// ListAccounts returns one page of accounts.
func (s *serviceImpl) ListAccounts(ctx context.Context, in ListAccountsInput) (*ListAccountsOutput, error) {
	pageNum, pageSize := normalizePage(in.PageNum, in.PageSize)
	model := dao.Account.Ctx(ctx)
	if name := strings.TrimSpace(in.Name); name != "" {
		model = model.WhereLike(dao.Account.Columns().Name, "%"+name+"%")
	}
	if in.Status != nil {
		model = model.Where(dao.Account.Columns().Status, *in.Status)
	}
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	var list []*entity.Account
	if err = model.Page(pageNum, pageSize).OrderDesc(dao.Account.Columns().Id).Scan(&list); err != nil {
		return nil, err
	}
	if list == nil {
		list = []*entity.Account{}
	}
	return &ListAccountsOutput{List: list, Total: total}, nil
}

// GetAccount returns one account by ID.
func (s *serviceImpl) GetAccount(ctx context.Context, id int64) (*entity.Account, error) {
	if id <= 0 {
		return nil, bizerr.NewCode(CodeAccountNotFound)
	}
	var row entity.Account
	err := dao.Account.Ctx(ctx).Where(dao.Account.Columns().Id, id).Scan(&row)
	if err != nil {
		return nil, err
	}
	if row.Id == 0 {
		return nil, bizerr.NewCode(CodeAccountNotFound)
	}
	return &row, nil
}

// CreateAccount creates one account with optional outbound/inbound bindings.
func (s *serviceImpl) CreateAccount(ctx context.Context, in CreateAccountInput) (int64, error) {
	if strings.TrimSpace(in.Name) == "" {
		return 0, bizerr.NewCode(CodeAccountNameRequired)
	}
	if err := s.validateAccountBindings(ctx, in.OutboundConnectionID, in.InboundConnectionID); err != nil {
		return 0, err
	}
	status := in.Status
	if status != StatusDisabled {
		status = StatusEnabled
	}
	isDefault := FlagNotDefault
	if in.IsDefault {
		isDefault = FlagDefault
		if _, err := dao.Account.Ctx(ctx).
			Where(dao.Account.Columns().IsDefault, FlagDefault).
			Data(do.Account{IsDefault: FlagNotDefault}).
			Update(); err != nil {
			return 0, err
		}
	}
	id, err := dao.Account.Ctx(ctx).Data(do.Account{
		Name:                 strings.TrimSpace(in.Name),
		FromAddress:          strings.TrimSpace(in.FromAddress),
		OutboundConnectionId: in.OutboundConnectionID,
		InboundConnectionId:  in.InboundConnectionID,
		IsDefault:            isDefault,
		Status:               status,
		TenantId:             0,
		Remark:               strings.TrimSpace(in.Remark),
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateAccount updates one account.
func (s *serviceImpl) UpdateAccount(ctx context.Context, in UpdateAccountInput) error {
	if _, err := s.GetAccount(ctx, in.ID); err != nil {
		return err
	}
	if strings.TrimSpace(in.Name) == "" {
		return bizerr.NewCode(CodeAccountNameRequired)
	}
	if err := s.validateAccountBindings(ctx, in.OutboundConnectionID, in.InboundConnectionID); err != nil {
		return err
	}
	status := in.Status
	if status != StatusDisabled {
		status = StatusEnabled
	}
	isDefault := FlagNotDefault
	if in.IsDefault {
		isDefault = FlagDefault
		if _, err := dao.Account.Ctx(ctx).
			WhereNot(dao.Account.Columns().Id, in.ID).
			Where(dao.Account.Columns().IsDefault, FlagDefault).
			Data(do.Account{IsDefault: FlagNotDefault}).
			Update(); err != nil {
			return err
		}
	}
	_, err := dao.Account.Ctx(ctx).Where(dao.Account.Columns().Id, in.ID).Data(do.Account{
		Name:                 strings.TrimSpace(in.Name),
		FromAddress:          strings.TrimSpace(in.FromAddress),
		OutboundConnectionId: in.OutboundConnectionID,
		InboundConnectionId:  in.InboundConnectionID,
		IsDefault:            isDefault,
		Status:               status,
		Remark:               strings.TrimSpace(in.Remark),
	}).Update()
	return err
}

// DeleteAccounts soft-deletes accounts by IDs.
func (s *serviceImpl) DeleteAccounts(ctx context.Context, ids []int64) error {
	ids = uniquePositiveIDs(ids)
	if len(ids) == 0 {
		return bizerr.NewCode(CodeIDsRequired)
	}
	_, err := dao.Account.Ctx(ctx).WhereIn(dao.Account.Columns().Id, ids).Delete()
	return err
}

// ResolveAccount resolves an explicit account or the platform default account.
func (s *serviceImpl) ResolveAccount(ctx context.Context, accountID int64) (*entity.Account, error) {
	if accountID > 0 {
		return s.GetAccount(ctx, accountID)
	}
	var row entity.Account
	err := dao.Account.Ctx(ctx).
		Where(dao.Account.Columns().IsDefault, FlagDefault).
		Where(dao.Account.Columns().Status, StatusEnabled).
		OrderDesc(dao.Account.Columns().Id).
		Limit(1).
		Scan(&row)
	if err != nil {
		return nil, err
	}
	if row.Id == 0 {
		return nil, bizerr.NewCode(mailcap.CodeMailAccountRequired)
	}
	return &row, nil
}

func (s *serviceImpl) validateAccountBindings(ctx context.Context, outboundID, inboundID int64) error {
	if outboundID > 0 {
		conn, err := s.GetConnection(ctx, outboundID)
		if err != nil {
			return err
		}
		if mailcap.Kind(conn.Kind) != mailcap.KindSMTP {
			return bizerr.NewCode(CodeAccountBindingInvalid)
		}
	}
	if inboundID > 0 {
		conn, err := s.GetConnection(ctx, inboundID)
		if err != nil {
			return err
		}
		kind := mailcap.Kind(conn.Kind)
		if kind != mailcap.KindIMAP && kind != mailcap.KindPOP3 {
			return bizerr.NewCode(CodeAccountBindingInvalid)
		}
	}
	return nil
}
