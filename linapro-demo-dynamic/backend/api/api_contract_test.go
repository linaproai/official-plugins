// This file keeps linapro-demo-dynamic API DTO boundary checks inside the plugin module.

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

// TestDemoDynamicAPIsDoNotDependOnGeneratedEntities ensures public API contracts do not import database entities.
func TestDemoDynamicAPIsDoNotDependOnGeneratedEntities(t *testing.T) {
	assertNoGeneratedEntityImports(t, demoDynamicPluginRoot(t))
}

// TestDemoDynamicAPIDTOsDoNotUseEntityNaming ensures response DTOs use API-oriented names.
func TestDemoDynamicAPIDTOsDoNotUseEntityNaming(t *testing.T) {
	assertNoEntityTypeNames(t, demoDynamicPluginRoot(t))
}

// TestDemoDynamicAPIDTOFilesAvoidLegacyNames rejects legacy DTO-only file naming.
func TestDemoDynamicAPIDTOFilesAvoidLegacyNames(t *testing.T) {
	assertNoLegacyDTOFiles(t, demoDynamicPluginRoot(t))
}

// TestDemoDynamicAPIDTOsKeepVersionPrefixInRouteGroup ensures DTO paths stay resource-local.
func TestDemoDynamicAPIDTOsKeepVersionPrefixInRouteGroup(t *testing.T) {
	assertNoAPIDTOPathPrefix(t, demoDynamicPluginRoot(t), "/api/v1")
}

// TestDemoDynamicAPIPackagesDoNotDeclareRouteGroupPrefix ensures route groups
// stay in backend route registration instead of generated API files.
func TestDemoDynamicAPIPackagesDoNotDeclareRouteGroupPrefix(t *testing.T) {
	assertNoRouteGroupPrefixConst(t, demoDynamicPluginRoot(t))
}

// TestDemoDynamicAPIDocI18NDoesNotReferenceRemovedDTOFields keeps apidoc translations aligned with DTOs.
func TestDemoDynamicAPIDocI18NDoesNotReferenceRemovedDTOFields(t *testing.T) {
	assertAPIDocI18NExcludesTokens(t, demoDynamicPluginRoot(t), removedAPIDocTokens())
}

// demoDynamicPluginRoot returns the plugin root directory for path-based contract checks.
func demoDynamicPluginRoot(t *testing.T) string {
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

// assertNoAPIDTOPathPrefix rejects version prefixes that belong to route groups.
func assertNoAPIDTOPathPrefix(t *testing.T, root string, forbiddenPrefix string) {
	t.Helper()

	for _, file := range collectAPIFiles(t, root) {
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			field, ok := node.(*ast.Field)
			if !ok || field == nil || field.Tag == nil || len(field.Names) != 0 {
				return true
			}
			tagValue := strings.Trim(field.Tag.Value, "`")
			pathValue := readStructTagValue(tagValue, "path")
			if strings.HasPrefix(pathValue, forbiddenPrefix) {
				t.Fatalf("dynamic plugin API DTO path in %s must not include route group prefix %s: %s", slashPath(root, file), forbiddenPrefix, pathValue)
			}
			return true
		})
	}
}

// assertNoRouteGroupPrefixConst rejects manual route group constants in API packages.
func assertNoRouteGroupPrefixConst(t *testing.T, root string) {
	t.Helper()

	for _, file := range collectAPIFiles(t, root) {
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			spec, ok := node.(*ast.ValueSpec)
			if !ok {
				return true
			}
			for _, name := range spec.Names {
				if name != nil && name.Name == "RouteGroupPrefix" {
					t.Fatalf("plugin API file %s must not declare RouteGroupPrefix; use backend RegisterPlugin Routes instead", slashPath(root, file))
				}
			}
			return true
		})
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

// readStructTagValue extracts one raw struct-tag key value without requiring a reflect.StructTag.
func readStructTagValue(tagValue string, key string) string {
	prefix := key + ":\""
	index := strings.Index(tagValue, prefix)
	if index < 0 {
		return ""
	}
	valueStart := index + len(prefix)
	cursor := valueStart
	for cursor < len(tagValue) {
		if tagValue[cursor] == '"' && tagValue[cursor-1] != '\\' {
			return tagValue[valueStart:cursor]
		}
		cursor++
	}
	return ""
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
