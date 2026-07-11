// This file defines the AI bridge wire codecs. Method request DTOs remain
// owned by the aicap subpackages; this package only serializes them for
// owner-aware dynamic host-service calls.

package bridge

import (
	"encoding/json"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// MarshalRequest encodes one owner AI request DTO as compact JSON.
func MarshalRequest(value any) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, gerror.Wrap(err, "encode ai bridge request failed")
	}
	return payload, nil
}

// UnmarshalRequest decodes one owner AI request DTO from compact JSON.
func UnmarshalRequest[T any](payload []byte) (*T, error) {
	if len(payload) == 0 {
		return nil, gerror.New("ai bridge request is empty")
	}
	out := new(T)
	if err := json.Unmarshal(payload, out); err != nil {
		return nil, gerror.Wrap(err, "decode ai bridge request failed")
	}
	return out, nil
}

// MarshalValueRequest wraps one owner AI request DTO in the generic JSON value
// envelope used by status-style host-service methods.
func MarshalValueRequest(value any) ([]byte, error) {
	payload, err := MarshalRequest(value)
	if err != nil {
		return nil, err
	}
	return protocol.MarshalHostServiceJSONRequest(&protocol.HostServiceJSONRequest{Value: payload}), nil
}

// UnmarshalValueRequest decodes one owner AI request DTO from the generic JSON
// value envelope.
func UnmarshalValueRequest[T any](payload []byte) (*T, error) {
	request, err := protocol.UnmarshalHostServiceJSONRequest(payload)
	if err != nil {
		return nil, err
	}
	if request == nil || len(request.Value) == 0 {
		return nil, gerror.New("ai bridge value request is empty")
	}
	return UnmarshalRequest[T](request.Value)
}

// MarshalResponse wraps one owner AI response DTO in the generic JSON value
// envelope returned by successful host-service calls.
func MarshalResponse(value any) ([]byte, error) {
	payload, err := MarshalRequest(value)
	if err != nil {
		return nil, err
	}
	return protocol.MarshalHostServiceJSONResponse(&protocol.HostServiceJSONResponse{Value: payload}), nil
}

// UnmarshalResponse decodes one owner AI response DTO from the generic JSON
// value envelope returned by the host.
func UnmarshalResponse[T any](payload []byte) (*T, error) {
	out := new(T)
	if err := DecodeResponse(payload, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DecodeResponse decodes one generic JSON value response into out.
func DecodeResponse(payload []byte, out any) error {
	if out == nil {
		return nil
	}
	response, err := protocol.UnmarshalHostServiceJSONResponse(payload)
	if err != nil {
		return err
	}
	if response == nil || len(response.Value) == 0 {
		return gerror.New("ai bridge response is empty")
	}
	if err = json.Unmarshal(response.Value, out); err != nil {
		return gerror.Wrap(err, "decode ai bridge response failed")
	}
	return nil
}
