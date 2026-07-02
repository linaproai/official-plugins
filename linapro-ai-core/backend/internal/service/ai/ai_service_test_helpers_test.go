// This file contains test-only helpers for Smart Center service integration tests.

package ai

import (
	"context"

	"lina-core/pkg/bizerr"
	"lina-plugin-linapro-ai-core/backend/internal/dao"
	"lina-plugin-linapro-ai-core/backend/internal/model/do"
	"lina-plugin-linapro-ai-core/backend/internal/model/entity"
)

// enabledEndpointForProtocol returns the first enabled provider endpoint for
// tests that need to assert endpoint selection setup.
func (s *serviceImpl) enabledEndpointForProtocol(ctx context.Context, providerID int64, protocol string) (*entity.ProviderEndpoint, error) {
	protocol = normalizeProtocol(protocol)
	if providerID <= 0 || protocol == "" {
		return nil, bizerr.NewCode(CodeProviderProtocolRequired)
	}
	var row *entity.ProviderEndpoint
	if err := dao.ProviderEndpoint.Ctx(ctx).
		Where(do.ProviderEndpoint{
			ProviderId: providerID,
			Protocol:   protocol,
			Enabled:    enabledYes,
		}).
		OrderAsc(dao.ProviderEndpoint.Columns().Id).
		Scan(&row); err != nil {
		return nil, err
	}
	if row == nil {
		return nil, bizerr.NewCode(CodeProviderProtocolRequired)
	}
	return row, nil
}

// modelCapabilityKeysByModel loads persisted capability identities for a model
// set in one query.
func (s *serviceImpl) modelCapabilityKeysByModel(ctx context.Context, modelIDs []int64) (map[string]struct{}, error) {
	rows, err := s.modelCapabilitiesByModel(ctx, modelIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		result[modelCapabilityKey(row.ModelId, row.CapabilityType, row.CapabilityMethod)] = struct{}{}
	}
	return result, nil
}

// insertConfirmedRemoteModelCapabilities inserts only capabilities explicitly
// returned by the provider.
func (s *serviceImpl) insertConfirmedRemoteModelCapabilities(
	ctx context.Context,
	existing map[string]struct{},
	model *entity.Model,
	endpointID int64,
	capabilities []remoteModelCapability,
) error {
	if model == nil || len(capabilities) == 0 {
		return nil
	}
	if existing == nil {
		existing = make(map[string]struct{})
	}
	for _, capability := range capabilities {
		capabilityType := normalizeCapabilityType(capability.CapabilityType)
		capabilityMethod := normalizeCapabilityMethod(capability.CapabilityMethod)
		if capabilityType == "" || capabilityMethod == "" {
			continue
		}
		key := modelCapabilityKey(model.Id, capabilityType, capabilityMethod)
		if _, ok := existing[key]; ok {
			continue
		}
		if _, err := dao.ModelCapability.Ctx(ctx).Data(do.ModelCapability{
			ModelId:           model.Id,
			EndpointId:        endpointID,
			CapabilityType:    capabilityType,
			CapabilityMethod:  capabilityMethod,
			InputModalities:   joinCSV(capability.InputModalities),
			OutputModalities:  joinCSV(capability.OutputModalities),
			SupportsOperation: enabledNo,
			SupportsStreaming: enabledNo,
			SupportsThinking:  enabledNo,
			Enabled:           enabledYes,
		}).Insert(); err != nil {
			return err
		}
		existing[key] = struct{}{}
	}
	return nil
}
