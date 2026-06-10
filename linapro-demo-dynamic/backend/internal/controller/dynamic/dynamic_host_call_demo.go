// Host call demo route controller.

package dynamic

import (
	"context"
	"strings"

	v1 "lina-plugin-linapro-demo-dynamic/backend/api/dynamic/v1"
	dynamicservice "lina-plugin-linapro-demo-dynamic/backend/internal/service/dynamic"

	bridgeplugin "lina-core/pkg/plugin/pluginbridge"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// HostCallDemo demonstrates unified host service capabilities including runtime,
// governed storage, outbound HTTP, and structured data access.
func (c *Controller) HostCallDemo(
	ctx context.Context,
	req *v1.HostCallDemoReq,
) (res *v1.HostCallDemoRes, err error) {
	payload, err := c.dynamicSvc.BuildHostCallDemoPayload(
		ctx,
		buildHostCallDemoRouteInput(bridgeplugin.RequestEnvelopeFromContext(ctx), req),
	)
	if err != nil {
		return nil, err
	}
	return &v1.HostCallDemoRes{
		VisitCount: payload.VisitCount,
		PluginID:   payload.PluginID,
		Runtime: &v1.HostCallDemoRuntimeRes{
			Now:  payload.Runtime.Now,
			UUID: payload.Runtime.UUID,
			Node: payload.Runtime.Node,
		},
		Storage: &v1.HostCallDemoStorageRes{
			PathPrefix:  payload.Storage.PathPrefix,
			ObjectPath:  payload.Storage.ObjectPath,
			Stored:      payload.Storage.Stored,
			ListedCount: payload.Storage.ListedCount,
			Deleted:     payload.Storage.Deleted,
		},
		Network: &v1.HostCallDemoNetworkRes{
			URL:         payload.Network.URL,
			Skipped:     payload.Network.Skipped,
			StatusCode:  payload.Network.StatusCode,
			ContentType: payload.Network.ContentType,
			BodyPreview: payload.Network.BodyPreview,
			Error:       payload.Network.Error,
		},
		Data: &v1.HostCallDemoDataRes{
			Table:      payload.Data.Table,
			RecordKey:  payload.Data.RecordKey,
			ListTotal:  payload.Data.ListTotal,
			CountTotal: payload.Data.CountTotal,
			Updated:    payload.Data.Updated,
			Deleted:    payload.Data.Deleted,
		},
		Config: &v1.HostCallDemoConfigRes{
			Plugin: &v1.HostCallDemoPluginConfigRes{
				Greeting:            payload.Config.Plugin.Greeting,
				GreetingFound:       payload.Config.Plugin.GreetingFound,
				FeatureEnabled:      payload.Config.Plugin.FeatureEnabled,
				FeatureEnabledFound: payload.Config.Plugin.FeatureEnabledFound,
			},
			HostConfig: &v1.HostCallDemoHostConfigRes{
				WorkspaceBasePath:      payload.Config.HostConfig.WorkspaceBasePath,
				WorkspaceBasePathFound: payload.Config.HostConfig.WorkspaceBasePathFound,
				I18nDefault:            payload.Config.HostConfig.I18nDefault,
				I18nDefaultFound:       payload.Config.HostConfig.I18nDefaultFound,
				I18nEnabled:            payload.Config.HostConfig.I18nEnabled,
				I18nEnabledFound:       payload.Config.HostConfig.I18nEnabledFound,
			},
		},
		Manifest: &v1.HostCallDemoManifestRes{
			ProfilePath:       payload.Manifest.ProfilePath,
			ProfileFound:      payload.Manifest.ProfileFound,
			ProfileName:       payload.Manifest.ProfileName,
			ProfileTier:       payload.Manifest.ProfileTier,
			ProfileOwner:      payload.Manifest.ProfileOwner,
			ConfigPath:        payload.Manifest.ConfigPath,
			ConfigFound:       payload.Manifest.ConfigFound,
			ConfigBodyPreview: payload.Manifest.ConfigBodyPreview,
		},
		Org: &v1.HostCallDemoOrgRes{
			Available:            payload.Org.Available,
			CapabilityID:         payload.Org.CapabilityID,
			ActiveProvider:       payload.Org.ActiveProvider,
			Reason:               payload.Org.Reason,
			AssignmentCount:      payload.Org.AssignmentCount,
			CurrentUserDeptCount: payload.Org.CurrentUserDeptCount,
			CurrentUserPostCount: payload.Org.CurrentUserPostCount,
		},
		Tenant: &v1.HostCallDemoTenantRes{
			Available:       payload.Tenant.Available,
			CapabilityID:    payload.Tenant.CapabilityID,
			ActiveProvider:  payload.Tenant.ActiveProvider,
			Reason:          payload.Tenant.Reason,
			CurrentTenantID: payload.Tenant.CurrentTenantID,
			PlatformBypass:  payload.Tenant.PlatformBypass,
			UserTenantCount: payload.Tenant.UserTenantCount,
			Visible:         payload.Tenant.Visible,
		},
		Message: payload.Message,
	}, nil
}

// buildHostCallDemoRouteInput extracts host-call demo metadata and flags from
// the bridge request envelope.
func buildHostCallDemoRouteInput(
	request *protocol.BridgeRequestEnvelopeV1,
	req *v1.HostCallDemoReq,
) *dynamicservice.HostCallDemoInput {
	input := &dynamicservice.HostCallDemoInput{}
	if request == nil {
		if req != nil {
			input.SkipNetwork = req.SkipNetwork
		}
		return input
	}

	input.PluginID = strings.TrimSpace(request.PluginID)
	input.RequestID = strings.TrimSpace(request.RequestID)
	if request.Route != nil {
		input.RoutePath = strings.TrimSpace(request.Route.InternalPath)
	}
	if req != nil {
		input.SkipNetwork = req.SkipNetwork
	}
	if request.Identity != nil {
		input.Username = strings.TrimSpace(request.Identity.Username)
		input.UserID = int(request.Identity.UserID)
	}
	return input
}
