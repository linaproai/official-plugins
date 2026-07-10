// This file implements the dynamic guest AI service client. The client keeps
// owner-aware transport details local to the bridge package and exposes the
// same typed aicap service contracts used by source plugins.

package bridge

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/pluginbridge"
	"lina-plugin-linapro-ai-core/backend/cap/aicap"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aiaudio"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aitypes"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aidocument"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aiembedding"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aiimage"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aisafety"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aitext"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aivideo"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aivision"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/spi"
)

// HostServiceInvoker dispatches one owner-aware host-service call and returns
// the encoded success payload from the host.
type HostServiceInvoker func(
	owner string,
	service string,
	version string,
	method string,
	resourceRef string,
	table string,
	payload []byte,
) ([]byte, error)

// Client is the dynamic guest-side AI namespace client.
type Client struct {
	invoke HostServiceInvoker
}

// textClient implements the dynamic guest text AI service.
type textClient struct{ client *Client }

var (
	_ aicap.Service       = (*Client)(nil)
	_ aitext.Service      = (*textClient)(nil)
	_ aiimage.Service     = (*imageClient)(nil)
	_ aiembedding.Service = (*embeddingClient)(nil)
	_ aiaudio.Service     = (*audioClient)(nil)
	_ aivision.Service    = (*visionClient)(nil)
	_ aidocument.Service  = (*documentClient)(nil)
	_ aisafety.Service    = (*safetyClient)(nil)
	_ aivideo.Service     = (*videoClient)(nil)
)

// New creates the default WASI-backed AI bridge client.
func New() aicap.Service {
	return NewClient(pluginbridge.InvokeOwnerHostService)
}

// NewClient creates one AI bridge client with an explicit host-service
// invoker. Tests can pass a fake invoker without depending on WASI imports.
func NewClient(invoker HostServiceInvoker) *Client {
	return &Client{invoke: invoker}
}

// MethodStatuses reads method-level availability for AI sub capabilities.
func (c *Client) MethodStatuses(_ context.Context, input aicap.MethodStatusesInput) (*aicap.MethodStatusesResult, error) {
	var response aicap.MethodStatusesResult
	if err := c.callValueRequest(spi.MethodStatusesBatchGet, input, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// Text returns the governed text AI guest client.
func (c *Client) Text() aitext.Service {
	if c == nil {
		return aitext.NewUnavailable()
	}
	return &textClient{client: c}
}

// Image returns the governed image AI guest client.
func (c *Client) Image() aiimage.Service {
	if c == nil {
		return aiimage.New()
	}
	return &imageClient{subClient: newSubClient(c, aiimage.CapabilityImageV1, aiimage.CapabilityType)}
}

// Embedding returns the governed embedding AI guest client.
func (c *Client) Embedding() aiembedding.Service {
	if c == nil {
		return aiembedding.New()
	}
	return &embeddingClient{subClient: newSubClient(c, aiembedding.CapabilityEmbeddingV1, aiembedding.CapabilityType)}
}

// Audio returns the governed audio AI guest client.
func (c *Client) Audio() aiaudio.Service {
	if c == nil {
		return aiaudio.New()
	}
	return &audioClient{subClient: newSubClient(c, aiaudio.CapabilityAudioV1, aiaudio.CapabilityType)}
}

// Vision returns the governed vision AI guest client.
func (c *Client) Vision() aivision.Service {
	if c == nil {
		return aivision.New()
	}
	return &visionClient{subClient: newSubClient(c, aivision.CapabilityVisionV1, aivision.CapabilityType)}
}

// Document returns the governed document AI guest client.
func (c *Client) Document() aidocument.Service {
	if c == nil {
		return aidocument.New()
	}
	return &documentClient{subClient: newSubClient(c, aidocument.CapabilityDocumentV1, aidocument.CapabilityType)}
}

// Safety returns the governed safety AI guest client.
func (c *Client) Safety() aisafety.Service {
	if c == nil {
		return aisafety.New()
	}
	return &safetyClient{subClient: newSubClient(c, aisafety.CapabilitySafetyV1, aisafety.CapabilityType)}
}

// Video returns the governed video AI guest client.
func (c *Client) Video() aivideo.Service {
	if c == nil {
		return aivideo.New()
	}
	return &videoClient{subClient: newSubClient(c, aivideo.CapabilityVideoV1, aivideo.CapabilityType)}
}

// callRequest encodes a direct owner DTO request and decodes the host response.
func (c *Client) callRequest(method string, input any, out any) error {
	payload, err := MarshalRequest(input)
	if err != nil {
		return err
	}
	return c.callPayload(method, payload, out)
}

// callValueRequest encodes a JSON value-envelope request and decodes the host
// response.
func (c *Client) callValueRequest(method string, input any, out any) error {
	payload, err := MarshalValueRequest(input)
	if err != nil {
		return err
	}
	return c.callPayload(method, payload, out)
}

// callPayload invokes one owner-aware AI method using the configured transport.
func (c *Client) callPayload(method string, payload []byte, out any) error {
	if c == nil || c.invoke == nil {
		return gerror.New("linapro-ai-core bridge host-service invoker is nil")
	}
	responsePayload, err := c.invoke(
		spi.OwnerPluginID,
		spi.ServiceAI,
		spi.VersionV1,
		method,
		"",
		"",
		payload,
	)
	if err != nil {
		return err
	}
	return DecodeResponse(responsePayload, out)
}

// methodStatus reads one non-text method status through the batch status
// method and falls back to unavailable when transport fails.
func (c *Client) methodStatus(
	ctx context.Context,
	capabilityID string,
	capabilityType aitypes.CapabilityType,
	method aitypes.CapabilityMethod,
) aitypes.MethodStatus {
	if c == nil {
		return aitypes.UnavailableMethodStatus(capabilityID, capabilityType, method)
	}
	result, err := c.MethodStatuses(ctx, aicap.MethodStatusesInput{
		Methods: []aicap.MethodStatusQuery{{
			CapabilityType:   aicap.CapabilityType(capabilityType),
			CapabilityMethod: aicap.CapabilityMethod(method),
		}},
	})
	if err != nil || result == nil || len(result.Items) == 0 {
		return aitypes.UnavailableMethodStatus(capabilityID, capabilityType, method)
	}
	return result.Items[0]
}

// unavailableStatus returns a safe capability unavailable projection for
// synchronous guest status methods.
func unavailableStatus(capabilityID string) capmodel.CapabilityStatus {
	return aitypes.UnavailableStatus(capabilityID)
}
