// This file verifies login-log retention cleanup against plugin-owned tables.

package loginlog

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	_ "lina-core/pkg/dbdriver"

	"lina-plugin-linapro-monitor-loginlog/backend/internal/dao"
	"lina-plugin-linapro-monitor-loginlog/backend/internal/model/do"
)

var (
	loginLogInstallSQLOnce  sync.Once
	loginLogInstallSQLError error
)

// TestCleanupExpiredDeletesOnlyLogsOlderThanRetention verifies global retention
// cleanup deletes old login logs without deleting newer records.
func TestCleanupExpiredDeletesOnlyLogsOlderThanRetention(t *testing.T) {
	ctx := context.Background()
	prepareLoginLogDatabase(t, ctx)
	svc := New(nil, nil)

	const rollbackMessage = "rollback login-log retention cleanup test transaction"
	err := dao.Loginlog.Transaction(ctx, func(txCtx context.Context, _ gdb.TX) error {
		oldID := insertLoginLogFixture(t, txCtx, "retention-old", time.Now().AddDate(0, 0, -2))
		newID := insertLoginLogFixture(t, txCtx, "retention-new", time.Now())

		deleted, err := svc.CleanupExpired(txCtx, 1)
		if err != nil {
			t.Fatalf("cleanup expired login logs: %v", err)
		}
		if deleted < 1 {
			t.Fatalf("expected at least the old fixture log to be deleted, got %d", deleted)
		}
		if loginLogExists(t, txCtx, oldID) {
			t.Fatalf("expected old login log %d to be deleted", oldID)
		}
		if !loginLogExists(t, txCtx, newID) {
			t.Fatalf("expected new login log %d to remain", newID)
		}
		return gerror.New(rollbackMessage)
	})
	if err == nil || err.Error() != rollbackMessage {
		t.Fatalf("expected transaction rollback marker %q, got %v", rollbackMessage, err)
	}
}

// prepareLoginLogDatabase installs plugin SQL once for DB-backed service tests.
func prepareLoginLogDatabase(t *testing.T, ctx context.Context) {
	t.Helper()
	adapter, err := gcfg.NewAdapterContent(`
database:
  default:
    link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
    debug: false
`)
	if err != nil {
		t.Fatalf("create config adapter: %v", err)
	}
	g.Cfg().SetAdapter(adapter)

	loginLogInstallSQLOnce.Do(func() {
		sqlPaths, readErr := filepath.Glob(filepath.Clean("../../../../manifest/sql/*.sql"))
		if readErr != nil {
			loginLogInstallSQLError = readErr
			return
		}
		sort.Strings(sqlPaths)
		for _, sqlPath := range sqlPaths {
			content, readErr := os.ReadFile(sqlPath)
			if readErr != nil {
				loginLogInstallSQLError = readErr
				return
			}
			if strings.TrimSpace(string(content)) == "" {
				continue
			}
			if _, execErr := g.DB().Exec(ctx, string(content)); execErr != nil {
				loginLogInstallSQLError = execErr
				return
			}
		}
	})
	if loginLogInstallSQLError != nil {
		t.Skipf("database unavailable for login-log service test: %v", loginLogInstallSQLError)
	}
	if _, err = g.DB().Exec(ctx, "SELECT 1"); err != nil {
		t.Skipf("database unavailable for login-log service test: %v", err)
	}
}

// insertLoginLogFixture creates one login-log row with a controlled timestamp.
func insertLoginLogFixture(t *testing.T, ctx context.Context, suffix string, loginTime time.Time) int64 {
	t.Helper()
	insertID, err := dao.Loginlog.Ctx(ctx).Data(do.Loginlog{
		UserName:  "unit-" + suffix,
		Status:    LoginStatusSuccess,
		Ip:        "127.0.0.1",
		Browser:   "unit",
		Os:        "unit",
		Msg:       "unit",
		LoginTime: &loginTime,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert login-log fixture: %v", err)
	}
	return int64(insertID)
}

// loginLogExists reports whether the fixture row still exists.
func loginLogExists(t *testing.T, ctx context.Context, id int64) bool {
	t.Helper()
	count, err := dao.Loginlog.Ctx(ctx).Where(do.Loginlog{Id: id}).Count()
	if err != nil {
		t.Fatalf("count login-log fixture: %v", err)
	}
	return count > 0
}
