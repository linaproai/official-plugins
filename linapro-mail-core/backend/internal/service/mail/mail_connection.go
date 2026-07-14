// This file implements Connection CRUD for linapro-mail-core.

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

// ListConnections returns one page of connections.
func (s *serviceImpl) ListConnections(ctx context.Context, in ListConnectionsInput) (*ListConnectionsOutput, error) {
	pageNum, pageSize := normalizePage(in.PageNum, in.PageSize)
	model := dao.Connection.Ctx(ctx)
	if name := strings.TrimSpace(in.Name); name != "" {
		model = model.WhereLike(dao.Connection.Columns().Name, "%"+name+"%")
	}
	if kind := strings.TrimSpace(in.Kind); kind != "" {
		model = model.Where(dao.Connection.Columns().Kind, kind)
	}
	if in.Status != nil {
		model = model.Where(dao.Connection.Columns().Status, *in.Status)
	}
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	var list []*entity.Connection
	if err = model.Page(pageNum, pageSize).OrderDesc(dao.Connection.Columns().Id).Scan(&list); err != nil {
		return nil, err
	}
	if list == nil {
		list = []*entity.Connection{}
	}
	return &ListConnectionsOutput{List: list, Total: total}, nil
}

// GetConnection returns one connection by ID.
func (s *serviceImpl) GetConnection(ctx context.Context, id int64) (*entity.Connection, error) {
	if id <= 0 {
		return nil, bizerr.NewCode(mailcap.CodeMailConnectionNotFound)
	}
	var row entity.Connection
	err := dao.Connection.Ctx(ctx).Where(dao.Connection.Columns().Id, id).Scan(&row)
	if err != nil {
		return nil, err
	}
	if row.Id == 0 {
		return nil, bizerr.NewCode(mailcap.CodeMailConnectionNotFound)
	}
	return &row, nil
}

// CreateConnection creates one connection.
func (s *serviceImpl) CreateConnection(ctx context.Context, in CreateConnectionInput) (int64, error) {
	if err := validateConnectionInput(in.Name, in.Kind, in.Host, in.Port, in.TLSMode, in.AuthMode); err != nil {
		return 0, err
	}
	status := in.Status
	if status != StatusDisabled {
		status = StatusEnabled
	}
	tlsMode := normalizeTLSMode(in.TLSMode)
	authMode := normalizeAuthMode(in.AuthMode)
	id, err := dao.Connection.Ctx(ctx).Data(do.Connection{
		Name:      strings.TrimSpace(in.Name),
		Kind:      strings.TrimSpace(in.Kind),
		Host:      strings.TrimSpace(in.Host),
		Port:      in.Port,
		Username:  strings.TrimSpace(in.Username),
		SecretRef: strings.TrimSpace(in.SecretRef),
		TlsMode:   tlsMode,
		AuthMode:  authMode,
		ExtraJson: defaultJSON(in.ExtraJSON),
		Status:    status,
		TenantId:  0,
		Remark:    strings.TrimSpace(in.Remark),
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateConnection updates one connection.
func (s *serviceImpl) UpdateConnection(ctx context.Context, in UpdateConnectionInput) error {
	if _, err := s.GetConnection(ctx, in.ID); err != nil {
		return err
	}
	if err := validateConnectionInput(in.Name, in.Kind, in.Host, in.Port, in.TLSMode, in.AuthMode); err != nil {
		return err
	}
	status := in.Status
	if status != StatusDisabled {
		status = StatusEnabled
	}
	_, err := dao.Connection.Ctx(ctx).Where(dao.Connection.Columns().Id, in.ID).Data(do.Connection{
		Name:      strings.TrimSpace(in.Name),
		Kind:      strings.TrimSpace(in.Kind),
		Host:      strings.TrimSpace(in.Host),
		Port:      in.Port,
		Username:  strings.TrimSpace(in.Username),
		SecretRef: strings.TrimSpace(in.SecretRef),
		TlsMode:   normalizeTLSMode(in.TLSMode),
		AuthMode:  normalizeAuthMode(in.AuthMode),
		ExtraJson: defaultJSON(in.ExtraJSON),
		Status:    status,
		Remark:    strings.TrimSpace(in.Remark),
	}).Update()
	return err
}

// DeleteConnections soft-deletes connections by IDs.
func (s *serviceImpl) DeleteConnections(ctx context.Context, ids []int64) error {
	ids = uniquePositiveIDs(ids)
	if len(ids) == 0 {
		return bizerr.NewCode(CodeIDsRequired)
	}
	_, err := dao.Connection.Ctx(ctx).WhereIn(dao.Connection.Columns().Id, ids).Delete()
	return err
}

func validateConnectionInput(name, kind, host string, port int, tlsMode, authMode string) error {
	if strings.TrimSpace(name) == "" {
		return bizerr.NewCode(CodeConnectionNameRequired)
	}
	switch mailcap.Kind(strings.TrimSpace(kind)) {
	case mailcap.KindSMTP, mailcap.KindIMAP, mailcap.KindPOP3:
	default:
		return bizerr.NewCode(CodeConnectionKindInvalid)
	}
	if strings.TrimSpace(host) == "" {
		return bizerr.NewCode(CodeConnectionHostRequired)
	}
	if port <= 0 || port > 65535 {
		return bizerr.NewCode(CodeConnectionPortInvalid)
	}
	_ = tlsMode
	_ = authMode
	return nil
}

func normalizeTLSMode(value string) string {
	switch strings.TrimSpace(value) {
	case TLSModeDisable, TLSModeTLS, TLSModeStartTLS:
		return strings.TrimSpace(value)
	default:
		return TLSModeStartTLS
	}
}

func normalizeAuthMode(value string) string {
	switch strings.TrimSpace(value) {
	case AuthModePassword:
		return strings.TrimSpace(value)
	default:
		return AuthModePassword
	}
}

func defaultJSON(value string) string {
	if strings.TrimSpace(value) == "" {
		return "{}"
	}
	return value
}
