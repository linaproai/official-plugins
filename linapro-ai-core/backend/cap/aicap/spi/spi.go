// Package spi publishes provider-side helper APIs for the linapro-ai-core
// owner capability. It turns typed AI provider factories into generic
// capability descriptors that the host can govern without importing AI
// business DTOs through pluginhost.
package spi

import (
	"lina-core/pkg/plugin/capability/capregistry"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aitext"
)

const (
	// OwnerPluginID identifies the plugin that owns the AI capability contract.
	OwnerPluginID = aitext.ProviderPluginID
	// ServiceAI is the owner-aware dynamic host service key.
	ServiceAI = "ai"
	// VersionV1 is the current AI owner capability protocol version.
	VersionV1 = "v1"
	// CapabilityAIV1 identifies the root plugin-owned AI capability namespace.
	CapabilityAIV1 = "plugin.linapro-ai-core.ai.v1"
	// SourceContract is the source-plugin import path for the AI capability.
	SourceContract = "lina-plugin-linapro-ai-core/backend/cap/aicap"
	// DynamicContract is the dynamic-plugin bridge SDK import path.
	DynamicContract = "lina-plugin-linapro-ai-core/backend/cap/aicap/bridge"
)

const (
	// MethodTextGenerate executes synchronous text generation.
	MethodTextGenerate = "text.generate"
	// MethodTextStatusGet reads text-generation method availability.
	MethodTextStatusGet = "text.method_status.get"
	// MethodStatusesBatchGet reads multiple AI method availability projections.
	MethodStatusesBatchGet = "ai.methods.status.batch_get"

	// Multimodal method constants remain stable for guest SDKs and future
	// provider bindings. They are intentionally not published by
	// MethodDescriptors until a real provider path is wired.
	// MethodImageGenerate executes image generation.
	MethodImageGenerate = "image.generate"
	// MethodImageEdit executes image editing.
	MethodImageEdit = "image.edit"
	// MethodEmbeddingCreate creates vector embeddings.
	MethodEmbeddingCreate = "embedding.create"
	// MethodAudioTranscribe executes speech-to-text transcription.
	MethodAudioTranscribe = "audio.transcribe"
	// MethodAudioSynthesize executes text-to-speech synthesis.
	MethodAudioSynthesize = "audio.synthesize"
	// MethodVisionAnalyze executes image, screenshot, or diagram analysis.
	MethodVisionAnalyze = "vision.analyze"
	// MethodDocumentAnalyze executes document analysis.
	MethodDocumentAnalyze = "document.analyze"
	// MethodDocumentCite executes citation-aware document analysis.
	MethodDocumentCite = "document.cite"
	// MethodSafetyModerate executes content safety moderation.
	MethodSafetyModerate = "safety.moderate"
	// MethodVideoGenerate starts video generation.
	MethodVideoGenerate = "video.generate"
	// MethodVideoEdit starts video editing.
	MethodVideoEdit = "video.edit"
	// MethodVideoExtend starts video extension.
	MethodVideoExtend = "video.extend"
	// MethodVideoOperationGet reads provider operation status.
	MethodVideoOperationGet = "video.operation.get"
	// MethodVideoOperationCancel cancels a provider operation.
	MethodVideoOperationCancel = "video.operation.cancel"
)

type (
	// ProviderEnv is the typed construction environment for AI providers.
	ProviderEnv = aitext.ProviderEnv
	// TextProviderFactory creates the text AI provider implementation.
	TextProviderFactory = aitext.ProviderFactory
)

// Descriptor returns the generic owner descriptor metadata for ai.v1 without a
// runtime invoker. Call aicap.ProviderDescriptor to attach the invoker.
func Descriptor() capregistry.Descriptor {
	return capregistry.Descriptor{
		OwnerPluginID:   OwnerPluginID,
		Service:         ServiceAI,
		Version:         VersionV1,
		SourceContract:  SourceContract,
		DynamicContract: DynamicContract,
		Methods:         MethodDescriptors(),
	}
}

// MethodDescriptors returns the AI methods currently published by the owner
// descriptor. Only methods with a real runtime path are published so
// authorization and catalog projections match actual host behavior.
// Multimodal DTO packages remain available for source consumption and will be
// added here when their provider SPI is wired.
func MethodDescriptors() []capregistry.MethodDescriptor {
	return []capregistry.MethodDescriptor{
		executeMethod(MethodTextGenerate, aitext.CapabilityTextV1, "aitext.GenerateRequest", "aitext.GenerateResponse"),
		readMethod(MethodTextStatusGet, aitext.CapabilityTextV1, "aicap.MethodStatusQuery", "aitypes.MethodStatus"),
		readMethod(MethodStatusesBatchGet, CapabilityAIV1, "aicap.MethodStatusesInput", "aicap.MethodStatusesResult"),
	}
}

func readMethod(method string, capability string, request string, response string) capregistry.MethodDescriptor {
	return ownerMethod(method, capability, capregistry.RiskLevelRead, capregistry.ResourceKindKey, request, response)
}

func executeMethod(method string, capability string, request string, response string) capregistry.MethodDescriptor {
	return ownerMethod(method, capability, capregistry.RiskLevelExecute, capregistry.ResourceKindKey, request, response)
}

func ownerMethod(
	method string,
	capability string,
	risk capregistry.RiskLevel,
	resourceKind capregistry.ResourceKind,
	request string,
	response string,
) capregistry.MethodDescriptor {
	return capregistry.MethodDescriptor{
		Method:          method,
		Capability:      capability,
		Risk:            risk,
		ResourceKind:    resourceKind,
		RequestPayload:  request,
		ResponsePayload: response,
	}
}
