package pluginstorageoss

import (
	"encoding/json"
	"io/fs"
	"testing"
)

func TestEmbeddedMenuJSONLocalizesSettingsTitle(t *testing.T) {
	t.Parallel()
	cases := []struct {
		path  string
		title string
	}{
		{path: "manifest/i18n/en-US/menu.json", title: "Alibaba Cloud OSS"},
		{path: "manifest/i18n/zh-CN/menu.json", title: "\u963f\u91cc\u4e91 OSS"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()
			raw, err := fs.ReadFile(EmbeddedFiles, tc.path)
			if err != nil {
				t.Fatal(err)
			}
			var payload struct {
				Menu map[string]struct {
					Title string `json:"title"`
				} `json:"menu"`
			}
			if err := json.Unmarshal(raw, &payload); err != nil {
				t.Fatal(err)
			}
			if payload.Menu["plugin:linapro-storage-oss:settings"].Title != tc.title {
				t.Fatalf("got %q want %q", payload.Menu["plugin:linapro-storage-oss:settings"].Title, tc.title)
			}
		})
	}
}
