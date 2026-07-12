package settings

import (
	"context"
	"testing"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/hostconfigcap"
)

type fakeSysConfig struct {
	values map[hostconfigcap.SysConfigKey]string
}

func newFakeSysConfig() *fakeSysConfig {
	return &fakeSysConfig{values: map[hostconfigcap.SysConfigKey]string{}}
}

func (f *fakeSysConfig) Get(_ context.Context, key hostconfigcap.SysConfigKey) (*hostconfigcap.SysConfigInfo, error) {
	if value, ok := f.values[key]; ok {
		return &hostconfigcap.SysConfigInfo{Key: key, Value: value}, nil
	}
	return nil, nil
}

func (f *fakeSysConfig) BatchGet(_ context.Context, keys []hostconfigcap.SysConfigKey) (*capmodel.BatchResult[*hostconfigcap.SysConfigInfo, hostconfigcap.SysConfigKey], error) {
	out := &capmodel.BatchResult[*hostconfigcap.SysConfigInfo, hostconfigcap.SysConfigKey]{
		Items: map[hostconfigcap.SysConfigKey]*hostconfigcap.SysConfigInfo{},
	}
	for _, key := range keys {
		if value, ok := f.values[key]; ok {
			out.Items[key] = &hostconfigcap.SysConfigInfo{Key: key, Value: value}
		} else {
			out.MissingIDs = append(out.MissingIDs, key)
		}
	}
	return out, nil
}

func (f *fakeSysConfig) List(context.Context, hostconfigcap.ListSysConfigInput) (*capmodel.PageResult[*hostconfigcap.SysConfigInfo], error) {
	return &capmodel.PageResult[*hostconfigcap.SysConfigInfo]{}, nil
}

func (f *fakeSysConfig) SetValue(_ context.Context, key hostconfigcap.SysConfigKey, value string) error {
	f.values[key] = value
	return nil
}

func (f *fakeSysConfig) Reset(context.Context, hostconfigcap.SysConfigKey) error { return nil }

func (f *fakeSysConfig) EnsureVisible(context.Context, []hostconfigcap.SysConfigKey) error {
	return nil
}

func TestSaveMasksSecretAndKeepsOnEmpty(t *testing.T) {
	t.Parallel()
	fake := newFakeSysConfig()
	svc := New(fake)
	ctx := context.Background()

	proj, err := svc.Save(ctx, SaveInput{
		AccessKeyID:     "ak-1",
		SecretAccessKey: "secret-1",
		Region:          "us-east-1",
		Bucket:          "demo",
	})
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if !proj.SecretAccessKeyConfigured || proj.SecretAccessKeyMasked != SecretMask {
		t.Fatalf("expected masked secret projection, got %#v", proj)
	}

	if _, err = svc.Save(ctx, SaveInput{
		AccessKeyID: "ak-1", Region: "us-east-1", Bucket: "demo",
	}); err != nil {
		t.Fatalf("save keep: %v", err)
	}
	snap, err := svc.Load(ctx)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if snap.SecretAccessKey != "secret-1" {
		t.Fatalf("expected secret kept, got %q", snap.SecretAccessKey)
	}
}

func TestValidateReadyRequiresAWSFields(t *testing.T) {
	t.Parallel()
	svc := New(newFakeSysConfig())
	if err := svc.ValidateReady(&Snapshot{AccessKeyID: "a"}); err == nil {
		t.Fatal("expected invalid config")
	}
	if err := svc.ValidateReady(&Snapshot{
		AccessKeyID: "a", SecretAccessKey: "s", Region: "us-east-1", Bucket: "b",
	}); err != nil {
		t.Fatalf("expected ready, got %v", err)
	}
}
