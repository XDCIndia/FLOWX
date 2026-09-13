package postgres

import (
	"context"
	"strings"
	"testing"

	"github.com/fluxa/fluxa/internal/domain"
	"github.com/fluxa/fluxa/internal/tenant"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// execCapturingDB stubs the DB interface and records the last Exec call.
type execCapturingDB struct {
	sql         string
	args        []interface{}
	commandTag  pgconn.CommandTag
	err         error
	execCounter int
}

func (db *execCapturingDB) Exec(_ context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	db.execCounter++
	db.sql = sql
	db.args = args
	return db.commandTag, db.err
}

func (db *execCapturingDB) Query(_ context.Context, _ string, _ ...interface{}) (pgx.Rows, error) {
	return nil, nil
}

func (db *execCapturingDB) QueryRow(_ context.Context, _ string, _ ...interface{}) pgx.Row {
	return nil
}

func (db *execCapturingDB) Begin(_ context.Context) (pgx.Tx, error) { return nil, nil }

// Regression (IDOR): WalletRepo.Delete must scope the DELETE by the tenant ID
// from context, mirroring GetByID. Without the predicate any tenant could
// delete any other tenant's wallet by ID.
func TestWalletRepoDelete_ScopesByTenant(t *testing.T) {
	db := &execCapturingDB{commandTag: pgconn.NewCommandTag("DELETE 1")}
	repo := NewWalletRepo(db)

	ctx := tenant.WithID(context.Background(), "tenant-a")
	if err := repo.Delete(ctx, "wallet-1"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	if !strings.Contains(db.sql, "AND tenant_id = $2") {
		t.Fatalf("Delete() SQL missing tenant predicate: %q", db.sql)
	}
	if len(db.args) != 2 || db.args[1] != "tenant-a" {
		t.Fatalf("Delete() args = %v, want [wallet-1 tenant-a]", db.args)
	}
}

// System calls (no tenant in context) must keep working: no tenant predicate,
// global delete by ID only.
func TestWalletRepoDelete_SystemContextHasNoTenantPredicate(t *testing.T) {
	db := &execCapturingDB{commandTag: pgconn.NewCommandTag("DELETE 1")}
	repo := NewWalletRepo(db)

	if err := repo.Delete(context.Background(), "wallet-1"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	if strings.Contains(db.sql, "tenant_id") {
		t.Fatalf("Delete() SQL unexpectedly scoped without tenant in context: %q", db.sql)
	}
	if len(db.args) != 1 || db.args[0] != "wallet-1" {
		t.Fatalf("Delete() args = %v, want [wallet-1]", db.args)
	}
}

// Tenant B attempting to delete tenant A's wallet matches zero rows; the repo
// must report the wallet as not found and must not silently succeed.
func TestWalletRepoDelete_CrossTenantDeleteMatchesNoRows(t *testing.T) {
	db := &execCapturingDB{commandTag: pgconn.NewCommandTag("DELETE 0")}
	repo := NewWalletRepo(db)

	ctx := tenant.WithID(context.Background(), "tenant-b")
	err := repo.Delete(ctx, "wallet-owned-by-tenant-a")
	if err != domain.ErrWalletNotFound {
		t.Fatalf("Delete() = %v, want ErrWalletNotFound (tenant B must not delete tenant A's wallet)", err)
	}
	if !strings.Contains(db.sql, "AND tenant_id = $2") {
		t.Fatalf("Delete() SQL missing tenant predicate: %q", db.sql)
	}
}
