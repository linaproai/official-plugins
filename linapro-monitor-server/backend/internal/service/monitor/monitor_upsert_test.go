// This file verifies linapro-monitor-server snapshot persistence against the
// supported PostgreSQL database path. Database-backed assertions are skipped
// unless LINA_TEST_PGSQL_LINK is explicitly provided.

package monitor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/dbdriver"
	"lina-core/pkg/dialect"
	"lina-plugin-linapro-monitor-server/backend/internal/dao"
)

// TestPostgreSQLDriverRegisteredForMonitorTests verifies the monitor tests can
// initialize GoFrame's supported PostgreSQL driver without external services.
func TestPostgreSQLDriverRegisteredForMonitorTests(t *testing.T) {
	driver, ok := dbdriver.New(dbdriver.TypePostgreSQL)
	if !ok || driver == nil {
		t.Fatalf("expected PostgreSQL driver factory to return a driver, got driver=%v ok=%t", driver, ok)
	}
}

// TestUpsertMonitorSnapshotWorksOnPostgreSQL verifies runtime persistence uses
// explicit Save conflict columns for monitor snapshots.
func TestUpsertMonitorSnapshotWorksOnPostgreSQL(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLMonitorServerDB(t, ctx)

	const (
		nodeName   = "unit-node"
		nodeIP     = "10.0.0.10"
		firstData  = `{"sample":1}`
		secondData = `{"sample":2}`
	)

	if err := upsertMonitorSnapshot(ctx, nodeName, nodeIP, firstData); err != nil {
		t.Fatalf("insert PostgreSQL monitor snapshot failed: %v", err)
	}
	if err := upsertMonitorSnapshot(ctx, nodeName, nodeIP, secondData); err != nil {
		t.Fatalf("update PostgreSQL monitor snapshot failed: %v", err)
	}

	count, err := dao.Server.Ctx(ctx).Count()
	if err != nil {
		t.Fatalf("count PostgreSQL monitor snapshots failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one snapshot row after duplicate node upsert, got %d", count)
	}

	var row *serverMonitorRecord
	if err = dao.Server.Ctx(ctx).Where(colNodeName, nodeName).Scan(&row); err != nil {
		t.Fatalf("read PostgreSQL monitor snapshot failed: %v", err)
	}
	if row == nil {
		t.Fatal("expected PostgreSQL monitor snapshot row to exist")
	}
	if row.Data != secondData {
		t.Fatalf("expected latest monitor data %s, got %s", secondData, row.Data)
	}
}

// TestGetDBInfoReturnsPostgreSQLVersion verifies linapro-monitor-server
// database diagnostics return a non-empty PostgreSQL version label.
func TestGetDBInfoReturnsPostgreSQLVersion(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLMonitorServerDB(t, ctx)

	info := New().GetDBInfo(ctx)
	if info == nil {
		t.Fatal("expected PostgreSQL DB info to be returned")
	}
	if !strings.Contains(strings.ToLower(info.Version), "postgresql") {
		t.Fatalf("expected PostgreSQL database version label, got %q", info.Version)
	}
	if strings.TrimSpace(info.Version) == "" {
		t.Fatal("expected PostgreSQL database version to be non-empty")
	}
}

// setupPostgreSQLMonitorServerDB points the generated DAO at a temporary
// PostgreSQL database and creates the linapro-monitor-server table.
func setupPostgreSQLMonitorServerDB(t *testing.T, ctx context.Context) {
	t.Helper()

	baseLink := strings.TrimSpace(os.Getenv("LINA_TEST_PGSQL_LINK"))
	if baseLink == "" {
		t.Skip("set LINA_TEST_PGSQL_LINK to run PostgreSQL monitor integration tests")
	}

	dbLink := postgresMonitorServerTestDatabaseLink(t, baseLink)
	dbDialect, err := dialect.From(dbLink)
	if err != nil {
		t.Fatalf("resolve PostgreSQL dialect failed: %v", err)
	}
	if err = dbDialect.PrepareDatabase(ctx, dbLink, true); err != nil {
		t.Fatalf("prepare PostgreSQL monitor database failed: %v", err)
	}
	t.Cleanup(func() {
		if cleanupErr := dropPostgreSQLMonitorServerTestDatabase(ctx, dbLink); cleanupErr != nil {
			t.Errorf("cleanup PostgreSQL monitor database failed: %v", cleanupErr)
		}
	})

	originalConfig := gdb.GetAllConfig()
	if err := gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: dbLink}},
	}); err != nil {
		t.Fatalf("configure PostgreSQL monitor database failed: %v", err)
	}
	db := g.DB()
	t.Cleanup(func() {
		if closeErr := db.Close(ctx); closeErr != nil {
			t.Errorf("close PostgreSQL monitor database failed: %v", closeErr)
		}
		if err := gdb.SetConfig(originalConfig); err != nil {
			t.Errorf("restore GoFrame database config failed: %v", err)
		}
	})

	sqlPath := filepath.Join("..", "..", "..", "..", "manifest", "sql", "001-linapro-monitor-server-schema.sql")
	content, err := os.ReadFile(sqlPath)
	if err != nil {
		t.Fatalf("read linapro-monitor-server schema SQL failed: %v", err)
	}
	translated, err := dbDialect.TranslateDDL(ctx, sqlPath, string(content))
	if err != nil {
		t.Fatalf("translate linapro-monitor-server schema SQL failed: %v", err)
	}
	for _, statement := range dialect.SplitSQLStatements(translated) {
		if _, err = db.Exec(ctx, statement); err != nil {
			t.Fatalf("execute linapro-monitor-server schema SQL failed: %v\nSQL:\n%s", err, statement)
		}
	}
}

// postgresMonitorServerTestDatabaseLink returns a unique database link for one
// monitor integration test.
func postgresMonitorServerTestDatabaseLink(t *testing.T, baseLink string) string {
	t.Helper()

	db, err := gdb.New(gdb.ConfigNode{Link: baseLink})
	if err != nil {
		t.Fatalf("parse PostgreSQL monitor base link failed: %v", err)
	}
	config := db.GetConfig()
	if config == nil {
		t.Fatal("PostgreSQL monitor base link configuration is empty")
	}
	if closeErr := db.Close(context.Background()); closeErr != nil {
		t.Fatalf("close PostgreSQL monitor base link parser failed: %v", closeErr)
	}

	extra := strings.TrimSpace(config.Extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/linapro_monitor_server_%d%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		time.Now().UnixNano(),
		extra,
	)
}

// dropPostgreSQLMonitorServerTestDatabase removes the temporary database used
// by one monitor integration test.
func dropPostgreSQLMonitorServerTestDatabase(ctx context.Context, targetLink string) (err error) {
	targetDB, err := gdb.New(gdb.ConfigNode{Link: targetLink})
	if err != nil {
		return err
	}
	targetConfig := targetDB.GetConfig()
	if targetConfig == nil {
		if closeErr := targetDB.Close(ctx); closeErr != nil {
			return closeErr
		}
		return nil
	}
	targetName := strings.TrimSpace(targetConfig.Name)
	if closeErr := targetDB.Close(ctx); closeErr != nil {
		return closeErr
	}
	if targetName == "" {
		return nil
	}

	systemDB, err := gdb.New(gdb.ConfigNode{Link: postgresMonitorServerSystemLink(*targetConfig)})
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := systemDB.Close(ctx); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	if _, err = systemDB.Exec(
		ctx,
		"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname=$1 AND pid<>pg_backend_pid()",
		targetName,
	); err != nil {
		return err
	}
	quotedName := `"` + strings.ReplaceAll(targetName, `"`, `""`) + `"`
	if _, err = systemDB.Exec(ctx, "DROP DATABASE IF EXISTS "+quotedName); err != nil {
		return err
	}
	return nil
}

// postgresMonitorServerSystemLink returns a PostgreSQL maintenance database
// link using the same host, credentials, and extra parameters.
func postgresMonitorServerSystemLink(config gdb.ConfigNode) string {
	extra := strings.TrimSpace(config.Extra)
	if extra != "" && !strings.HasPrefix(extra, "?") {
		extra = "?" + extra
	}
	return fmt.Sprintf(
		"pgsql:%s:%s@%s(%s:%s)/postgres%s",
		config.User,
		config.Pass,
		config.Protocol,
		config.Host,
		config.Port,
		extra,
	)
}
