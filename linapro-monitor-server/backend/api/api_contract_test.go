// This file keeps linapro-monitor-server API DTO boundary checks inside the plugin module.

package api

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestMonitorServerAPIsDoNotDependOnGeneratedEntities ensures public API contracts do not import database entities.
func TestMonitorServerAPIsDoNotDependOnGeneratedEntities(t *testing.T) {
	assertNoGeneratedEntityImports(t, monitorServerPluginRoot(t))
}

// TestMonitorServerAPIDTOsDoNotUseEntityNaming ensures response DTOs use API-oriented names.
func TestMonitorServerAPIDTOsDoNotUseEntityNaming(t *testing.T) {
	assertNoEntityTypeNames(t, monitorServerPluginRoot(t))
}

// TestMonitorServerAPIDTOFilesAvoidLegacyNames rejects legacy DTO-only file naming.
func TestMonitorServerAPIDTOFilesAvoidLegacyNames(t *testing.T) {
	assertNoLegacyDTOFiles(t, monitorServerPluginRoot(t))
}

// TestMonitorServerAPIDocI18NDoesNotReferenceRemovedDTOFields keeps apidoc translations aligned with DTOs.
func TestMonitorServerAPIDocI18NDoesNotReferenceRemovedDTOFields(t *testing.T) {
	assertAPIDocI18NExcludesTokens(t, monitorServerPluginRoot(t), removedAPIDocTokens())
}

// TestMonitorServerJobI18NResourcesMatchHandlerRefs keeps job handler i18n keys aligned with host handler refs.
func TestMonitorServerJobI18NResourcesMatchHandlerRefs(t *testing.T) {
	root := monitorServerPluginRoot(t)
	assertJobI18NValue(
		t,
		root,
		"en-US",
		"job.handler.plugin.linapro-monitor-server.jobs.service-monitor-collector.name",
		"Server Monitor Collection",
	)
	assertJobI18NValue(
		t,
		root,
		"zh-CN",
		"job.handler.plugin.linapro-monitor-server.jobs.service-monitor-collector.name",
		"服务监控采集",
	)
	assertJobI18NValue(
		t,
		root,
		"en-US",
		"job.handler.plugin.linapro-monitor-server.jobs.service-monitor-cleanup.name",
		"Server Monitor Cleanup",
	)
	assertJobI18NValue(
		t,
		root,
		"zh-CN",
		"job.handler.plugin.linapro-monitor-server.jobs.service-monitor-cleanup.name",
		"服务监控清理",
	)
}

// monitorServerPluginRoot returns the plugin root directory for path-based contract checks.
func monitorServerPluginRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

// assertNoGeneratedEntityImports rejects API source imports from generated entity packages.
func assertNoGeneratedEntityImports(t *testing.T, root string) {
	t.Helper()

	for _, file := range collectAPIFiles(t, root) {
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		for _, imported := range parsed.Imports {
			path := strings.Trim(imported.Path.Value, `"`)
			if strings.Contains(path, "/internal/model/entity") {
				t.Fatalf("plugin API file %s imports generated entity package %s", slashPath(root, file), path)
			}
		}
	}
}

// assertNoEntityTypeNames rejects response DTO type names that look like database entities.
func assertNoEntityTypeNames(t *testing.T, root string) {
	t.Helper()

	for _, file := range collectAPIFiles(t, root) {
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			spec, ok := node.(*ast.TypeSpec)
			if !ok {
				return true
			}
			if strings.HasSuffix(spec.Name.Name, "Entity") {
				t.Fatalf("plugin API type %s in %s must use response DTO naming instead of Entity", spec.Name.Name, slashPath(root, file))
			}
			return true
		})
	}
}

// assertNoLegacyDTOFiles rejects old DTO/entity sidecar files in API packages.
func assertNoLegacyDTOFiles(t *testing.T, root string) {
	t.Helper()

	for _, file := range collectAPIFiles(t, root) {
		name := filepath.Base(file)
		if strings.HasSuffix(name, "_entity.go") || strings.HasSuffix(name, "_dto.go") {
			t.Fatalf("plugin API DTO file %s must be folded into the API main source file", slashPath(root, file))
		}
	}
}

// assertAPIDocI18NExcludesTokens verifies plugin apidoc bundles no longer reference removed DTO names or fields.
func assertAPIDocI18NExcludesTokens(t *testing.T, root string, tokens []string) {
	t.Helper()

	for _, file := range collectAPIDocI18NFiles(t, root) {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		if !json.Valid(content) {
			t.Fatalf("plugin apidoc i18n file %s is not valid JSON", slashPath(root, file))
		}
		for _, token := range tokens {
			if bytes.Contains(content, []byte(token)) {
				t.Fatalf("plugin apidoc i18n file %s still references removed DTO token %q", slashPath(root, file), token)
			}
		}
	}
}

// assertJobI18NValue verifies one runtime job translation key in this plugin's manifest resources.
func assertJobI18NValue(t *testing.T, root string, locale string, key string, expected string) {
	t.Helper()

	file := filepath.Join(root, "manifest", "i18n", locale, "job.json")
	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read %s: %v", slashPath(root, file), err)
	}
	var payload map[string]any
	if err = json.Unmarshal(content, &payload); err != nil {
		t.Fatalf("parse %s: %v", slashPath(root, file), err)
	}
	actual, ok := nestedJSONString(payload, strings.Split(key, "."))
	if !ok {
		t.Fatalf("plugin job i18n file %s missing string key %q", slashPath(root, file), key)
	}
	if actual != expected {
		t.Fatalf("plugin job i18n key %q in %s: got %q want %q", key, slashPath(root, file), actual, expected)
	}
}

// nestedJSONString resolves a dot-separated JSON object path and returns a string leaf.
func nestedJSONString(payload map[string]any, parts []string) (string, bool) {
	var current any = payload
	for _, part := range parts {
		object, ok := current.(map[string]any)
		if !ok {
			return "", false
		}
		current, ok = object[part]
		if !ok {
			return "", false
		}
	}
	value, ok := current.(string)
	return value, ok
}

// collectAPIFiles lists non-test Go source files under this plugin's backend API directories.
func collectAPIFiles(t *testing.T, root string) []string {
	t.Helper()

	var files []string
	apiRoot := filepath.Join(root, "backend", "api")
	if err := filepath.WalkDir(apiRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk plugin API files: %v", err)
	}
	return files
}

// collectAPIDocI18NFiles lists plugin apidoc translation JSON files.
func collectAPIDocI18NFiles(t *testing.T, root string) []string {
	t.Helper()

	var files []string
	i18nRoot := filepath.Join(root, "manifest", "i18n")
	if err := filepath.WalkDir(i18nRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel := slashPath(i18nRoot, path)
		if strings.Contains(rel, "/apidoc/") && strings.HasSuffix(rel, ".json") {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk plugin apidoc i18n files: %v", err)
	}
	return files
}

// removedAPIDocTokens returns legacy schema names and fields removed from plugin apidoc resources.
func removedAPIDocTokens() []string {
	return []string{
		"NoticeEntity",
		"DeptEntity",
		"PostEntity",
		"LoginLogEntity",
		"OperLogEntity",
		"TenantEntity",
		"LoginTenantEntity",
		"TenantPluginEntity",
		"deletedAt",
	}
}

// slashPath returns a stable slash-separated path relative to root.
func slashPath(root string, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}
