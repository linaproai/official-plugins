// This file keeps linapro-tenant-core API DTO boundary checks inside the plugin module.

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
	"reflect"
	"runtime"
	"strings"
	"testing"

	authv1 "lina-plugin-linapro-tenant-core/backend/api/auth/v1"
	platformv1 "lina-plugin-linapro-tenant-core/backend/api/platform/v1"
	tenantv1 "lina-plugin-linapro-tenant-core/backend/api/tenant/v1"
)

// TestTenantCoreAPIsDoNotDependOnGeneratedEntities ensures public API contracts do not import database entities.
func TestTenantCoreAPIsDoNotDependOnGeneratedEntities(t *testing.T) {
	assertNoGeneratedEntityImports(t, tenantCorePluginRoot(t))
}

// TestTenantCoreAPIDTOsDoNotUseEntityNaming ensures response DTOs use API-oriented names.
func TestTenantCoreAPIDTOsDoNotUseEntityNaming(t *testing.T) {
	assertNoEntityTypeNames(t, tenantCorePluginRoot(t))
}

// TestTenantCoreAPIDTOFilePlacement keeps shared DTOs in the API package main source file.
func TestTenantCoreAPIDTOFilePlacement(t *testing.T) {
	assertTypeDeclaredInFile(t, tenantCorePluginRoot(t), map[string]string{
		"auth/v1/LoginTenantItem":    "auth.go",
		"platform/v1/TenantItem":     "platform.go",
		"tenant/v1/TenantPluginItem": "tenant.go",
	})
}

// TestTenantCoreAPIDTOFilesAvoidLegacyNames rejects legacy DTO-only file naming.
func TestTenantCoreAPIDTOFilesAvoidLegacyNames(t *testing.T) {
	assertNoLegacyDTOFiles(t, tenantCorePluginRoot(t))
}

// TestTenantCoreResponseDTOsHideInternalFields verifies implementation fields stay out of responses.
func TestTenantCoreResponseDTOsHideInternalFields(t *testing.T) {
	assertJSONFieldsAbsent(t, "auth.LoginTenantItem", reflect.TypeOf(authv1.LoginTenantItem{}), internalResponseFields())
	assertJSONFieldsAbsent(t, "platform.TenantItem", reflect.TypeOf(platformv1.TenantItem{}), internalResponseFields())
	assertJSONFieldsAbsent(t, "tenant.TenantPluginItem", reflect.TypeOf(tenantv1.TenantPluginItem{}), internalResponseFields())
}

// TestTenantCoreAPIDocI18NDoesNotReferenceRemovedDTOFields keeps apidoc translations aligned with DTOs.
func TestTenantCoreAPIDocI18NDoesNotReferenceRemovedDTOFields(t *testing.T) {
	assertAPIDocI18NExcludesTokens(t, tenantCorePluginRoot(t), []string{
		"TenantEntity",
		"LoginTenantEntity",
		"TenantPluginEntity",
		"deletedAt",
	})
}

// tenantCorePluginRoot returns the plugin root directory for path-based contract checks.
func tenantCorePluginRoot(t *testing.T) string {
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

// assertTypeDeclaredInFile verifies shared response DTOs stay in their main API source files.
func assertTypeDeclaredInFile(t *testing.T, root string, expected map[string]string) {
	t.Helper()

	seen := make(map[string]string, len(expected))
	for _, file := range collectAPIFiles(t, root) {
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		for _, declaration := range parsed.Decls {
			general, ok := declaration.(*ast.GenDecl)
			if !ok || general.Tok != token.TYPE {
				continue
			}
			for _, spec := range general.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				key := apiDTOKey(root, file, typeSpec.Name.Name)
				if _, ok := expected[key]; ok {
					seen[key] = filepath.Base(file)
				}
			}
		}
	}

	for key, wantFile := range expected {
		gotFile, ok := seen[key]
		if !ok {
			t.Fatalf("expected response DTO %s to exist", key)
		}
		if gotFile != wantFile {
			t.Fatalf("response DTO %s must be defined in %s, got %s", key, wantFile, gotFile)
		}
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

// apiDTOKey builds a stable API-relative key for a type declaration.
func apiDTOKey(root string, file string, typeName string) string {
	var (
		apiRoot = filepath.Join(root, "backend", "api")
		rel     = slashPath(apiRoot, file)
		dir     = strings.TrimSuffix(rel, "/"+filepath.Base(file))
	)
	return dir + "/" + typeName
}

// slashPath returns a stable slash-separated path relative to root.
func slashPath(root string, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}

// internalResponseFields returns implementation field names that must not leak through public response DTOs.
func internalResponseFields() []string {
	return []string{"password", "deletedAt", "path", "engine", "hash"}
}

// assertJSONFieldsAbsent verifies that a DTO does not publish the provided JSON fields.
func assertJSONFieldsAbsent(t *testing.T, name string, typ reflect.Type, fields []string) {
	t.Helper()

	jsonFields := jsonFieldSet(typ)
	for _, field := range fields {
		if jsonFields[field] {
			t.Fatalf("%s exposes internal field %q", name, field)
		}
	}
}

// jsonFieldSet collects JSON field names from a DTO, including embedded structs.
func jsonFieldSet(typ reflect.Type) map[string]bool {
	fields := make(map[string]bool)
	collectJSONFields(typ, fields)
	return fields
}

// collectJSONFields recursively records JSON field names for a struct type.
func collectJSONFields(typ reflect.Type, fields map[string]bool) {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if field.Anonymous {
			collectJSONFields(field.Type, fields)
			if name == "" {
				continue
			}
		}
		if name == "-" {
			continue
		}
		if name == "" {
			name = field.Name
		}
		fields[name] = true
	}
}
