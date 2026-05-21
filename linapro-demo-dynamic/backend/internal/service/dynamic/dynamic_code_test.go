// This file verifies dynamic-demo structured business error metadata.

package dynamicservice

import (
	"testing"
	"time"

	"lina-core/pkg/bizerr"
)

// TestDynamicDemoBusinessErrorMetadata verifies dynamic-demo errors expose
// stable runtime codes, i18n keys, and named message parameters.
func TestDynamicDemoBusinessErrorMetadata(t *testing.T) {
	err := bizerr.NewCode(CodeDynamicDemoRecordTitleTooLong, bizerr.P("maxChars", 128))
	messageErr, ok := bizerr.As(err)
	if !ok {
		t.Fatalf("expected structured business error, got %T", err)
	}
	if messageErr.RuntimeCode() != "PLUGIN_DEMO_DYNAMIC_RECORD_TITLE_TOO_LONG" {
		t.Fatalf("expected title-too-long code, got %q", messageErr.RuntimeCode())
	}
	if messageErr.MessageKey() != "error.plugin.demo.dynamic.record.title.too.long" {
		t.Fatalf("expected title-too-long key, got %q", messageErr.MessageKey())
	}
	params := messageErr.Params()
	if params["maxChars"] != 128 {
		t.Fatalf("expected maxChars message param 128, got %#v", params["maxChars"])
	}
}

// TestDemoRecordMilliFromStringParsesDatabaseTimestamp verifies demo-record
// timestamps are converted without the GoFrame parser path used outside WASM.
func TestDemoRecordMilliFromStringParsesDatabaseTimestamp(t *testing.T) {
	got := demoRecordMilliFromString("2026-04-16 09:00:00")
	if got == nil {
		t.Fatal("expected database timestamp string to parse")
	}

	want := time.Date(2026, 4, 16, 9, 0, 0, 0, time.Local).UnixMilli()
	if *got != want {
		t.Fatalf("expected timestamp millis %d, got %d", want, *got)
	}
	if demoRecordMilliFromString("not-a-time") != nil {
		t.Fatal("expected invalid timestamp string to project as nil")
	}
}

// TestDemoRecordMilliFromStringParsesTimestampOffset verifies database
// timestamp strings with explicit offsets do not rely on generic time parsers.
func TestDemoRecordMilliFromStringParsesTimestampOffset(t *testing.T) {
	got := demoRecordMilliFromString("2026-04-16T09:00:00.123456+08:30")
	if got == nil {
		t.Fatal("expected offset timestamp string to parse")
	}

	want := time.Date(2026, 4, 16, 9, 0, 0, 123456000, time.FixedZone("", 8*60*60+30*60)).UnixMilli()
	if *got != want {
		t.Fatalf("expected timestamp millis %d, got %d", want, *got)
	}
}
