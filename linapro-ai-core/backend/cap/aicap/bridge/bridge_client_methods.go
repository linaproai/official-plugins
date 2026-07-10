// This file contains the typed AI bridge method implementations. Each method
// uses the owner DTO from the matching aicap subpackage and invokes the
// owner-aware ai.v1 host service.

package bridge

import (
	"context"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-plugin-linapro-ai-core/backend/cap/aicap"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aiaudio"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aidocument"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aiembedding"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aiimage"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aisafety"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aitext"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aitypes"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aivideo"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aivision"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/spi"
)

// subClient shares synchronous guest status helpers for non-text modalities.
// Actual method availability is resolved through owner host-service status
// methods rather than local provider state.
type subClient struct {
	client         *Client
	capabilityID   string
	capabilityType aitypes.CapabilityType
}

// Available reports that synchronous guest status uses method-level calls.
func (*subClient) Available(context.Context) bool { return false }

// Status returns a safe AI fallback status for synchronous guest checks.
func (c *subClient) Status(context.Context) capmodel.CapabilityStatus {
	return unavailableStatus(c.capabilityID)
}

// MethodStatus reads method availability through the owner host service.
func (c *subClient) MethodStatus(ctx context.Context, method aitypes.CapabilityMethod) aitypes.MethodStatus {
	return c.client.methodStatus(ctx, c.capabilityID, c.capabilityType, method)
}

func newSubClient(client *Client, capabilityID string, capabilityType aitypes.CapabilityType) subClient {
	return subClient{
		client:         client,
		capabilityID:   capabilityID,
		capabilityType: capabilityType,
	}
}

// Available reports that synchronous guest status uses method-level calls.
func (*textClient) Available(context.Context) bool { return false }

// Status returns a safe text AI fallback status for synchronous guest checks.
func (*textClient) Status(context.Context) capmodel.CapabilityStatus {
	return unavailableStatus(aitext.CapabilityTextV1)
}

// MethodStatus reads text method availability through the owner host service.
func (c *textClient) MethodStatus(_ context.Context, method aitypes.CapabilityMethod) aitypes.MethodStatus {
	var response aitypes.MethodStatus
	err := c.client.callValueRequest(spi.MethodTextStatusGet, aicap.MethodStatusQuery{
		CapabilityType:   aicap.CapabilityTypeText,
		CapabilityMethod: aicap.CapabilityMethod(method),
	}, &response)
	if err != nil {
		return aitypes.UnavailableMethodStatus(aitext.CapabilityTextV1, aitypes.CapabilityTypeText, method)
	}
	return response
}

// GenerateText executes one governed text generation request.
func (c *textClient) GenerateText(_ context.Context, request aitext.GenerateRequest) (*aitext.GenerateResponse, error) {
	var response aitext.GenerateResponse
	if err := c.client.callRequest(spi.MethodTextGenerate, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

type imageClient struct{ subClient }

// Generate executes one governed image generation request.
func (c *imageClient) Generate(_ context.Context, request aiimage.GenerateRequest) (*aiimage.Response, error) {
	var response aiimage.Response
	if err := c.client.callRequest(spi.MethodImageGenerate, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// Edit executes one governed image editing request.
func (c *imageClient) Edit(_ context.Context, request aiimage.EditRequest) (*aiimage.Response, error) {
	var response aiimage.Response
	if err := c.client.callRequest(spi.MethodImageEdit, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

type embeddingClient struct{ subClient }

// Create executes one governed embedding creation request.
func (c *embeddingClient) Create(_ context.Context, request aiembedding.CreateRequest) (*aiembedding.CreateResponse, error) {
	var response aiembedding.CreateResponse
	if err := c.client.callRequest(spi.MethodEmbeddingCreate, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

type audioClient struct{ subClient }

// Transcribe executes one governed audio transcription request.
func (c *audioClient) Transcribe(_ context.Context, request aiaudio.TranscribeRequest) (*aiaudio.TranscribeResponse, error) {
	var response aiaudio.TranscribeResponse
	if err := c.client.callRequest(spi.MethodAudioTranscribe, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// Synthesize executes one governed audio synthesis request.
func (c *audioClient) Synthesize(_ context.Context, request aiaudio.SynthesizeRequest) (*aiaudio.SynthesizeResponse, error) {
	var response aiaudio.SynthesizeResponse
	if err := c.client.callRequest(spi.MethodAudioSynthesize, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

type visionClient struct{ subClient }

// Analyze executes one governed visual analysis request.
func (c *visionClient) Analyze(_ context.Context, request aivision.AnalyzeRequest) (*aivision.AnalyzeResponse, error) {
	var response aivision.AnalyzeResponse
	if err := c.client.callRequest(spi.MethodVisionAnalyze, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

type documentClient struct{ subClient }

// Analyze executes one governed document analysis request.
func (c *documentClient) Analyze(_ context.Context, request aidocument.AnalyzeRequest) (*aidocument.Response, error) {
	var response aidocument.Response
	if err := c.client.callRequest(spi.MethodDocumentAnalyze, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// Cite executes one governed citation-aware document request.
func (c *documentClient) Cite(_ context.Context, request aidocument.CiteRequest) (*aidocument.Response, error) {
	var response aidocument.Response
	if err := c.client.callRequest(spi.MethodDocumentCite, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

type safetyClient struct{ subClient }

// Moderate executes one governed safety moderation request.
func (c *safetyClient) Moderate(_ context.Context, request aisafety.ModerateRequest) (*aisafety.ModerateResponse, error) {
	var response aisafety.ModerateResponse
	if err := c.client.callRequest(spi.MethodSafetyModerate, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

type videoClient struct{ subClient }

// Generate executes one governed video generation request.
func (c *videoClient) Generate(_ context.Context, request aivideo.GenerateRequest) (*aivideo.Response, error) {
	var response aivideo.Response
	if err := c.client.callRequest(spi.MethodVideoGenerate, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// Edit executes one governed video editing request.
func (c *videoClient) Edit(_ context.Context, request aivideo.EditRequest) (*aivideo.Response, error) {
	var response aivideo.Response
	if err := c.client.callRequest(spi.MethodVideoEdit, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// Extend executes one governed video extension request.
func (c *videoClient) Extend(_ context.Context, request aivideo.ExtendRequest) (*aivideo.Response, error) {
	var response aivideo.Response
	if err := c.client.callRequest(spi.MethodVideoExtend, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// OperationGet reads one governed provider operation.
func (c *videoClient) OperationGet(_ context.Context, request aivideo.OperationGetRequest) (*aivideo.Response, error) {
	var response aivideo.Response
	if err := c.client.callRequest(spi.MethodVideoOperationGet, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// OperationCancel cancels one governed provider operation.
func (c *videoClient) OperationCancel(_ context.Context, request aivideo.OperationCancelRequest) (*aitypes.ProviderOperationRef, error) {
	var response aitypes.ProviderOperationRef
	if err := c.client.callRequest(spi.MethodVideoOperationCancel, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}
