// This file verifies the owner AI SPI helper publishes a generic descriptor
// that the core capability registry can validate and index.
package spi

import (
	"testing"

	"lina-core/pkg/plugin/capability/capregistry"
)

func TestDescriptorPublishesStableOwnerContracts(t *testing.T) {
	t.Parallel()

	descriptor := Descriptor()
	if descriptor.OwnerPluginID != OwnerPluginID ||
		descriptor.Service != ServiceAI ||
		descriptor.Version != VersionV1 ||
		descriptor.SourceContract != SourceContract ||
		descriptor.DynamicContract != DynamicContract {
		t.Fatalf("unexpected descriptor metadata: %#v", descriptor)
	}
	if descriptor.Invoker != nil {
		t.Fatal("pure Descriptor metadata must not attach a runtime invoker")
	}

	methods := make(map[string]capregistry.MethodDescriptor, len(descriptor.Methods))
	for _, method := range descriptor.Methods {
		if method.Method == "" ||
			method.Capability == "" ||
			method.RequestPayload == "" ||
			method.ResponsePayload == "" ||
			method.ResourceKind != capregistry.ResourceKindKey ||
			method.Risk == "" {
			t.Fatalf("method descriptor must publish complete governance metadata: %#v", method)
		}
		methods[method.Method] = method
	}
	for _, method := range []string{
		MethodTextGenerate,
		MethodTextStatusGet,
		MethodStatusesBatchGet,
	} {
		if _, ok := methods[method]; !ok {
			t.Fatalf("expected published method %s", method)
		}
	}
	for _, method := range []string{
		MethodImageGenerate,
		MethodVideoOperationCancel,
		"computer.act",
		"ui.operate",
	} {
		if _, ok := methods[method]; ok {
			t.Fatalf("method %s must stay outside the published AI owner descriptor", method)
		}
	}
	if methods[MethodTextGenerate].Risk != capregistry.RiskLevelExecute {
		t.Fatalf("expected text generation execute risk, got %#v", methods[MethodTextGenerate])
	}
	if methods[MethodTextStatusGet].Risk != capregistry.RiskLevelRead {
		t.Fatalf("expected text status read risk, got %#v", methods[MethodTextStatusGet])
	}
	if methods[MethodStatusesBatchGet].Risk != capregistry.RiskLevelRead {
		t.Fatalf("expected batch status read risk, got %#v", methods[MethodStatusesBatchGet])
	}
}
