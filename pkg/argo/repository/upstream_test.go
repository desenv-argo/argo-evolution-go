package argo_repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestLatestUpstreamSnapshotRestoresChanges(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, time.August, 19, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery(`(?s)SELECT \* FROM "argo_upstream_snapshots" ORDER BY checked_at DESC, created_at DESC LIMIT \$1`).WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "repository", "branch", "baseline_sha", "latest_sha", "latest_version", "behind_by", "status", "checked_at", "changes_json"}).
			AddRow("snapshot-1", "evolution-foundation/evolution-go", "main", "base", "latest", "0.8.0", 1, "update_available", now, `[{"sha":"abc1234","title":"fix: reconnect","category":"fix"}]`))
	snapshot, err := NewRepository(db).LatestUpstreamSnapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if snapshot == nil || len(snapshot.Changes) != 1 || snapshot.Changes[0].Category != "fix" {
		t.Fatalf("unexpected snapshot: %+v", snapshot)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
