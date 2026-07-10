// This file publishes the owner capability descriptor with a runtime invoker.
// Descriptor metadata stays in spi so pluginhost and governance code can read
// method tables without constructing provider factories.

package aicap

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/capregistry"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aitext"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/spi"
)

// ProviderDescriptor validates the typed text provider factory and returns the
// generic AI capability descriptor with a runtime invoker attached.
func ProviderDescriptor(factory aitext.ProviderFactory) (capregistry.Descriptor, error) {
	if factory == nil {
		return capregistry.Descriptor{}, gerror.New("linapro-ai-core: text provider factory is nil")
	}
	descriptor := spi.Descriptor()
	descriptor.Invoker = NewOwnerInvoker(factory)
	return descriptor, nil
}
