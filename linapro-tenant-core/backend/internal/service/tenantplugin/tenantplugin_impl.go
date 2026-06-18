// tenantplugin_impl.go implements tenant-plugin enablement orchestration
// through the host plugincap domain contract. Host plugin registry, tenant
// state, and cache revision tables stay owned by lina-core adapters.

package tenantplugin

import (
	"context"
	"strconv"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-plugin-linapro-tenant-core/backend/internal/service/shared"
)

// List returns tenant-controllable plugins with current tenant enablement.
func (s *serviceImpl) List(ctx context.Context) (*ListOutput, error) {
	tenantID, err := s.requireTenantID(ctx)
	if err != nil {
		return nil, err
	}
	if err = s.requirePlugincap(); err != nil {
		return nil, err
	}
	out, err := s.plugins.ListTenantPlugins(ctx, s.capabilityContext(ctx, tenantID, "tenant_plugin.list"), plugincap.TenantListInput{})
	if err != nil {
		return nil, err
	}
	list := make([]*Entity, 0)
	if out != nil {
		list = make([]*Entity, 0, len(out.Items))
		for _, item := range out.Items {
			if item == nil {
				continue
			}
			list = append(list, pluginEntity(item))
		}
	}
	return &ListOutput{List: list, Total: len(list)}, nil
}

// SetEnabled updates one tenant plugin enablement row.
func (s *serviceImpl) SetEnabled(ctx context.Context, pluginID string, enabled bool) error {
	tenantID, err := s.requireTenantID(ctx)
	if err != nil {
		return err
	}
	if err = s.requirePlugincap(); err != nil {
		return err
	}
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return bizerr.NewCode(CodePluginNotFound)
	}
	if !enabled && s.pluginLifecycleSvc != nil {
		if err = s.pluginLifecycleSvc.EnsureTenantPluginDisableAllowed(ctx, normalizedPluginID, int(tenantID)); err != nil {
			return err
		}
	}
	if err = s.pluginAdmin.SetEnabled(
		ctx,
		s.capabilityContext(ctx, tenantID, "tenant_plugin.set_enabled"),
		plugincap.PluginID(normalizedPluginID),
		enabled,
	); err != nil {
		return err
	}
	if !enabled && s.pluginLifecycleSvc != nil {
		s.pluginLifecycleSvc.NotifyTenantPluginDisabled(ctx, normalizedPluginID, int(tenantID))
	}
	return nil
}

// ProvisionForTenant provisions missing default tenant-scoped plugin enablement
// for a tenant. Existing tenant-owned enablement rows are preserved by the
// host plugincap owner.
func (s *serviceImpl) ProvisionForTenant(ctx context.Context, tenantID int64) error {
	if tenantID <= shared.PlatformTenantID {
		return nil
	}
	if err := s.requirePlugincap(); err != nil {
		return err
	}
	return s.pluginAdmin.ProvisionTenantDefaults(
		ctx,
		s.capabilityContext(ctx, tenantID, "tenant_plugin.provision"),
		plugincapTenantID(tenantID),
	)
}

// requireTenantID returns the current request tenant id or a stable bizerr.
func (s *serviceImpl) requireTenantID(ctx context.Context) (int64, error) {
	bizCtx := s.bizCtxSvc.Current(ctx)
	tenantID := int64(bizCtx.TenantID)
	if tenantID <= shared.PlatformTenantID {
		return 0, bizerr.NewCode(CodeTenantRequired)
	}
	return tenantID, nil
}

// pluginEntity converts plugincap's tenant projection into this plugin's API shape.
func pluginEntity(row *plugincap.TenantProjection) *Entity {
	if row == nil {
		return nil
	}
	return &Entity{
		Id:            string(row.ID),
		Name:          row.Name,
		Version:       row.Version,
		Type:          row.Type,
		Description:   row.Description,
		Installed:     boolInt(row.Installed),
		Enabled:       boolInt(row.Enabled),
		ScopeNature:   row.ScopeNature,
		InstallMode:   row.InstallMode,
		TenantEnabled: boolInt(row.TenantEnabled),
	}
}

// boolInt converts a boolean to the existing API's 0/1 response shape.
func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

// plugincapTenantID encodes one tenant identifier for plugincap.
func plugincapTenantID(tenantID int64) capmodel.DomainID {
	return capmodel.DomainID(strconv.FormatInt(tenantID, 10))
}
