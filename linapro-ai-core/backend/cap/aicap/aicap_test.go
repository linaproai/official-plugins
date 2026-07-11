// This file verifies the AI namespace fallback behavior, method status
// projections, and owner runtime i18n coverage for errors and authorization
// display labels.

package aicap

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aiaudio"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aiimage"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aitext"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aitypes"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/aivideo"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/spi"
)

func TestTextReturnsFallbackService(t *testing.T) {
	t.Parallel()

	service := New(nil)
	text := service.Text()
	if text == nil {
		t.Fatal("expected non-nil text AI fallback service")
	}
	_, err := text.GenerateText(context.Background(), aitext.GenerateRequest{
		Purpose: "test.summary",
		Tier:    aitext.TierStandard,
		Messages: []aitext.Message{
			{Role: aitext.MessageRoleUser, Content: "hello"},
		},
	})
	if !bizerr.Is(err, aitext.CodeTextProviderUnavailable) {
		t.Fatalf("expected provider unavailable fallback error, got %v", err)
	}
}

func TestForPluginReturnsFallbackService(t *testing.T) {
	t.Parallel()

	service := ForPlugin(nil, "source-plugin")
	if service == nil || service.Text() == nil || service.Image() == nil || service.Embedding() == nil ||
		service.Audio() == nil || service.Vision() == nil || service.Document() == nil ||
		service.Safety() == nil || service.Video() == nil {
		t.Fatal("expected plugin-scoped AI fallback service")
	}
}

func TestMultimodalFallbackServicesReturnUnavailable(t *testing.T) {
	t.Parallel()

	service := New(nil)
	_, err := service.Image().Generate(context.Background(), imageGenerateRequest())
	if !bizerr.Is(err, aitypes.CodeProviderUnavailable) {
		t.Fatalf("expected image provider unavailable, got %v", err)
	}
	_, err = service.Audio().Transcribe(context.Background(), audioTranscribeRequest())
	if !bizerr.Is(err, aitypes.CodeProviderUnavailable) {
		t.Fatalf("expected audio provider unavailable, got %v", err)
	}
	_, err = service.Video().OperationGet(context.Background(), videoOperationGetRequest())
	if !bizerr.Is(err, aitypes.CodeProviderUnavailable) {
		t.Fatalf("expected video operation provider unavailable, got %v", err)
	}
}

func TestMethodStatusesReturnUnavailableAndEnforceBatchLimit(t *testing.T) {
	t.Parallel()

	service := New(nil)
	result, err := service.MethodStatuses(context.Background(), MethodStatusesInput{Methods: []MethodStatusQuery{
		{
			CapabilityType:   CapabilityTypeText,
			CapabilityMethod: CapabilityMethod(aitypes.CapabilityMethodTextGenerate),
		},
		{
			CapabilityType:   CapabilityTypeImage,
			CapabilityMethod: CapabilityMethod(aitypes.CapabilityMethodImageGenerate),
		},
		{
			CapabilityType:   CapabilityType("unknown"),
			CapabilityMethod: CapabilityMethod("render"),
		},
	}})
	if err != nil {
		t.Fatalf("method statuses: %v", err)
	}
	if result == nil || len(result.Items) != 3 {
		t.Fatalf("expected three method statuses, got %#v", result)
	}
	for _, item := range result.Items {
		if item.Available || item.Reason == "" {
			t.Fatalf("expected unavailable status with reason, got %#v", item)
		}
	}
	if result.Items[0].CapabilityStatus.CapabilityID != aitext.CapabilityTextV1 {
		t.Fatalf("unexpected text capability status: %#v", result.Items[0])
	}
	if result.Items[1].CapabilityStatus.CapabilityID != aiimage.CapabilityImageV1 {
		t.Fatalf("unexpected image capability status: %#v", result.Items[1])
	}
	if result.Items[2].CapabilityStatus.CapabilityID != "" ||
		result.Items[2].CapabilityType != aitypes.CapabilityType("unknown") ||
		result.Items[2].CapabilityMethod != aitypes.CapabilityMethod("render") {
		t.Fatalf("unexpected unknown capability status: %#v", result.Items[2])
	}

	oversized := make([]MethodStatusQuery, MaxMethodStatusBatchSize+1)
	_, err = service.MethodStatuses(context.Background(), MethodStatusesInput{Methods: oversized})
	if !bizerr.Is(err, capmodel.CodeCapabilityLimitExceeded) {
		t.Fatalf("expected batch limit error, got %v", err)
	}
}

func TestAINamespaceDoesNotExposeWeakGateway(t *testing.T) {
	t.Parallel()

	serviceType := reflect.TypeOf((*Service)(nil)).Elem()
	for _, method := range []string{"Invoke", "Generate"} {
		if _, ok := serviceType.MethodByName(method); ok {
			t.Fatalf("AI namespace must not expose weak gateway method %s", method)
		}
	}
}

func TestRuntimeI18NCoversOwnerErrorsAndAuthorizationLabels(t *testing.T) {
	t.Parallel()

	requiredKeys := append(ownerErrorMessageKeys(), ownerAuthorizationMessageKeys()...)
	for _, locale := range []string{"en-US", "zh-CN"} {
		locale := locale
		t.Run(locale, func(t *testing.T) {
			t.Parallel()

			messages := loadPluginRuntimeMessages(t, locale)
			missing := make([]string, 0)
			for _, key := range requiredKeys {
				if strings.TrimSpace(messages[key]) == "" {
					missing = append(missing, key)
				}
			}
			if len(missing) > 0 {
				t.Fatalf("missing %s runtime i18n keys: %s", locale, strings.Join(missing, ", "))
			}
		})
	}
}

func imageGenerateRequest() aiimage.GenerateRequest {
	return aiimage.GenerateRequest{
		Purpose: "asset.preview",
		Tier:    aitypes.TierStandard,
		Prompt:  "draw a chart",
		Count:   1,
	}
}

func audioTranscribeRequest() aiaudio.TranscribeRequest {
	return aiaudio.TranscribeRequest{
		Purpose: "meeting.transcript",
		Tier:    aitypes.TierStandard,
		Audio: aitypes.AssetRef{
			Ref:       "asset/audio-1",
			MimeType:  "audio/mpeg",
			SizeBytes: 1024,
		},
	}
}

func videoOperationGetRequest() aivideo.OperationGetRequest {
	return aivideo.OperationGetRequest{
		Purpose:      "video.preview",
		OperationRef: "operation-1",
	}
}

func ownerErrorMessageKeys() []string {
	codes := []*bizerr.Code{
		aitypes.CodeProviderUnavailable,
		aitypes.CodePurposeRequired,
		aitypes.CodeTierInvalid,
		aitypes.CodeAssetRefRequired,
		aitypes.CodeOperationRefRequired,
		aitypes.CodeUnsupportedMethod,
		aitext.CodeTextProviderUnavailable,
		aitext.CodeTextTierInvalid,
		aitext.CodeTextMessagesRequired,
		aitext.CodeTextMessageRoleInvalid,
		aitext.CodeTextThinkingEffortInvalid,
		aitext.CodeTextPurposeRequired,
		aitext.CodeTextMetadataTooLarge,
		aitext.CodeTextMaxOutputTokensInvalid,
	}
	keys := make([]string, 0, len(codes))
	for _, code := range codes {
		keys = append(keys, code.MessageKey())
	}
	sort.Strings(keys)
	return keys
}

func ownerAuthorizationMessageKeys() []string {
	methods := spi.MethodDescriptors()
	keys := make([]string, 0, len(methods)+2)
	keys = append(
		keys,
		"plugin.linapro-ai-core.capability.services."+spi.ServiceAI,
		"plugin.linapro-ai-core.capability.versions."+spi.VersionV1,
	)
	for _, method := range methods {
		keys = append(keys, "plugin.linapro-ai-core.capability.methods."+method.Method)
	}
	sort.Strings(keys)
	return keys
}

func loadPluginRuntimeMessages(t *testing.T, locale string) map[string]string {
	t.Helper()

	pluginRoot := pluginRootFromCaller(t)
	messages := make(map[string]string)
	for _, file := range []string{"error.json", "plugin.json"} {
		path := filepath.Join(pluginRoot, "manifest", "i18n", locale, file)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read runtime i18n file %s: %v", path, err)
		}
		var payload any
		if err = json.Unmarshal(content, &payload); err != nil {
			t.Fatalf("parse runtime i18n file %s: %v", path, err)
		}
		flattenRuntimeI18N(messages, "", payload)
	}
	return messages
}

func pluginRootFromCaller(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", ".."))
}

func flattenRuntimeI18N(out map[string]string, prefix string, value any) {
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			trimmed := strings.TrimSpace(key)
			if trimmed == "" {
				continue
			}
			nextPrefix := trimmed
			if prefix != "" {
				nextPrefix = prefix + "." + trimmed
			}
			flattenRuntimeI18N(out, nextPrefix, nested)
		}
	case string:
		out[prefix] = typed
	}
}
