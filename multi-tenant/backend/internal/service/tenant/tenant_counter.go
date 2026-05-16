// This file implements tenant-counting dependencies for lifecycle preconditions.

package tenant

import (
	"context"

	"lina-plugin-multi-tenant/backend/internal/service/shared"
)

// ExistingCounter counts existing tenants without depending on CRUD collaborators.
type ExistingCounter struct{}

// CountExisting returns the number of non-deleted tenants.
func (c ExistingCounter) CountExisting(ctx context.Context) (int, error) {
	return shared.Model(ctx, shared.TableTenant).Count()
}
