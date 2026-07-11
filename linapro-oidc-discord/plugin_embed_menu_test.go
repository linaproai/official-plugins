// This file verifies that Discord OIDC menu i18n resources are embedded so the
// host can localize settings menu titles by Accept-Language.
package pluginoidcdiscord

import (
	"encoding/json"
	"io/fs"
	"testing"
)

// TestEmbeddedMenuJSONLocalizesSettingsTitle ensures en-US/zh-CN menu.json ship
// with the settings and save-settings keys expected by host menu projection.
func TestEmbeddedMenuJSONLocalizesSettingsTitle(t *testing.T) {
	t.Parallel()

	cases := []struct {
		path  string
		title string
	}{
		{
			path:  "manifest/i18n/en-US/menu.json",
			title: "Discord Login",
		},
		{
			path:  "manifest/i18n/zh-CN/menu.json",
			title: "Discord 登录",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()
			raw, err := fs.ReadFile(EmbeddedFiles, tc.path)
			if err != nil {
				t.Fatalf("read embedded menu resource: %v", err)
			}
			var payload struct {
				Menu map[string]struct {
					Title string `json:"title"`
				} `json:"menu"`
			}
			if err := json.Unmarshal(raw, &payload); err != nil {
				t.Fatalf("parse menu json: %v", err)
			}
			settings := payload.Menu["plugin:linapro-oidc-discord:settings"]
			if settings.Title != tc.title {
				t.Fatalf("settings title: got %q want %q", settings.Title, tc.title)
			}
			if payload.Menu["plugin:linapro-oidc-discord:settings-update"].Title == "" {
				t.Fatal("settings-update title must not be empty")
			}
		})
	}
}
