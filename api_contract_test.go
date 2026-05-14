// This file verifies source-plugin API response DTO boundaries.
package linaplugins

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

	noticev1 "lina-plugin-content-notice/backend/api/notice/v1"
	loginlogv1 "lina-plugin-monitor-loginlog/backend/api/loginlog/v1"
	operlogv1 "lina-plugin-monitor-operlog/backend/api/operlog/v1"
	authv1 "lina-plugin-multi-tenant/backend/api/auth/v1"
	platformv1 "lina-plugin-multi-tenant/backend/api/platform/v1"
	tenantv1 "lina-plugin-multi-tenant/backend/api/tenant/v1"
	deptv1 "lina-plugin-org-center/backend/api/dept/v1"
	postv1 "lina-plugin-org-center/backend/api/post/v1"
)

// TestPluginAPIsDoNotDependOnGeneratedEntities ensures API contracts are independent from database entities.
func TestPluginAPIsDoNotDependOnGeneratedEntities(t *testing.T) {
	root := pluginWorkspaceRoot(t)
	files := collectPluginAPIFiles(t, root)

	for _, file := range files {
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

// TestPluginAPIDTOsDoNotUseEntityNaming ensures API response DTOs are not named like database entities.
func TestPluginAPIDTOsDoNotUseEntityNaming(t *testing.T) {
	root := pluginWorkspaceRoot(t)
	files := collectPluginAPIFiles(t, root)

	for _, file := range files {
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

// TestPluginAPIDTOFilePlacement keeps shared response DTOs in each API package main source file.
func TestPluginAPIDTOFilePlacement(t *testing.T) {
	root := pluginWorkspaceRoot(t)
	expected := map[string]string{
		"content-notice/backend/api/notice/v1/NoticeItem":          "notice.go",
		"monitor-loginlog/backend/api/loginlog/v1/LoginLogItem":    "loginlog.go",
		"monitor-operlog/backend/api/operlog/v1/OperLogListItem":   "operlog.go",
		"monitor-operlog/backend/api/operlog/v1/OperLogDetailItem": "operlog.go",
		"multi-tenant/backend/api/auth/v1/LoginTenantItem":         "auth.go",
		"multi-tenant/backend/api/platform/v1/TenantItem":          "platform.go",
		"multi-tenant/backend/api/tenant/v1/TenantPluginItem":      "tenant.go",
		"org-center/backend/api/dept/v1/DeptItem":                  "dept.go",
		"org-center/backend/api/post/v1/PostItem":                  "post.go",
	}

	seen := make(map[string]string, len(expected))
	files := collectPluginAPIFiles(t, root)
	for _, file := range files {
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
				key := dtoKey(root, file, typeSpec.Name.Name)
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

// TestPluginAPIDTOFilesAvoidLegacyNames rejects legacy DTO-only file naming.
func TestPluginAPIDTOFilesAvoidLegacyNames(t *testing.T) {
	root := pluginWorkspaceRoot(t)
	files := collectPluginAPIFiles(t, root)

	for _, file := range files {
		name := filepath.Base(file)
		if strings.HasSuffix(name, "_entity.go") || strings.HasSuffix(name, "_dto.go") {
			t.Fatalf("plugin API DTO file %s must be folded into the API main source file", slashPath(root, file))
		}
	}
}

// TestPluginResponseDTOsHideInternalFields verifies sensitive implementation fields stay out of API responses.
func TestPluginResponseDTOsHideInternalFields(t *testing.T) {
	bannedFields := []string{"password", "deletedAt", "path", "engine", "hash"}
	dtoTypes := []struct {
		name string
		typ  reflect.Type
	}{
		{name: "notice.NoticeItem", typ: reflect.TypeOf(noticev1.NoticeItem{})},
		{name: "dept.DeptItem", typ: reflect.TypeOf(deptv1.DeptItem{})},
		{name: "post.PostItem", typ: reflect.TypeOf(postv1.PostItem{})},
		{name: "loginlog.LoginLogItem", typ: reflect.TypeOf(loginlogv1.LoginLogItem{})},
		{name: "operlog.OperLogListItem", typ: reflect.TypeOf(operlogv1.OperLogListItem{})},
		{name: "operlog.OperLogDetailItem", typ: reflect.TypeOf(operlogv1.OperLogDetailItem{})},
		{name: "auth.LoginTenantItem", typ: reflect.TypeOf(authv1.LoginTenantItem{})},
		{name: "platform.TenantItem", typ: reflect.TypeOf(platformv1.TenantItem{})},
		{name: "tenant.TenantPluginItem", typ: reflect.TypeOf(tenantv1.TenantPluginItem{})},
	}

	for _, dto := range dtoTypes {
		fields := jsonFieldSet(dto.typ)
		for _, field := range bannedFields {
			if fields[field] {
				t.Fatalf("%s exposes internal field %q", dto.name, field)
			}
		}
	}
}

// TestPluginOperLogListDTOExcludesPayloads keeps large audited payloads out of list responses.
func TestPluginOperLogListDTOExcludesPayloads(t *testing.T) {
	fields := jsonFieldSet(reflect.TypeOf(operlogv1.OperLogListItem{}))
	for _, field := range []string{"operParam", "jsonResult"} {
		if fields[field] {
			t.Fatalf("operlog.OperLogListItem exposes detail-only field %q", field)
		}
	}

	detailFields := jsonFieldSet(reflect.TypeOf(operlogv1.OperLogDetailItem{}))
	for _, field := range []string{"operParam", "jsonResult"} {
		if !detailFields[field] {
			t.Fatalf("operlog.OperLogDetailItem must expose audited payload field %q", field)
		}
	}
}

// TestPluginAPIDocI18NDoesNotReferenceRemovedDTOFields keeps apidoc translations aligned with response DTOs.
func TestPluginAPIDocI18NDoesNotReferenceRemovedDTOFields(t *testing.T) {
	root := pluginWorkspaceRoot(t)
	var files []string
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		rel := slashPath(root, path)
		if strings.Contains(rel, "/manifest/i18n/") && strings.Contains(rel, "/apidoc/") && strings.HasSuffix(rel, ".json") {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk plugin apidoc i18n files: %v", err)
	}

	banned := [][]byte{
		[]byte("NoticeEntity"),
		[]byte("DeptEntity"),
		[]byte("PostEntity"),
		[]byte("LoginLogEntity"),
		[]byte("OperLogEntity"),
		[]byte("TenantEntity"),
		[]byte("LoginTenantEntity"),
		[]byte("TenantPluginEntity"),
		[]byte("deletedAt"),
	}
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		if !json.Valid(content) {
			t.Fatalf("plugin apidoc i18n file %s is not valid JSON", slashPath(root, file))
		}
		for _, token := range banned {
			if bytes.Contains(content, token) {
				t.Fatalf("plugin apidoc i18n file %s still references removed DTO token %q", slashPath(root, file), token)
			}
		}
	}
}

// pluginWorkspaceRoot returns the lina-plugins module root for path-based contract checks.
func pluginWorkspaceRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file path")
	}
	return filepath.Dir(file)
}

// collectPluginAPIFiles lists Go source files under plugin backend API directories.
func collectPluginAPIFiles(t *testing.T, root string) []string {
	t.Helper()

	var files []string
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		rel := slashPath(root, path)
		if !strings.HasSuffix(rel, ".go") || strings.HasSuffix(rel, "_test.go") {
			return nil
		}
		if strings.Contains(rel, "/backend/api/") {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk plugin API files: %v", err)
	}

	return files
}

// dtoKey builds a stable key for a type declaration in a plugin API package.
func dtoKey(root string, file string, typeName string) string {
	rel := slashPath(root, file)
	dir := strings.TrimSuffix(rel, "/"+filepath.Base(file))
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
