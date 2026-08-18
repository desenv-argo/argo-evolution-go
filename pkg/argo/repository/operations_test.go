package argo_repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	argo_model "github.com/evolution-foundation/evolution-go/pkg/argo/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestGatewayUsageGroupsApplicationsAndInstances(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	from := time.Date(2026, time.August, 18, 0, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)
	columns := []string{"key", "total", "succeeded", "failed", "unverified_identity", "average_duration_ms", "last_activity_at"}
	mock.ExpectQuery(`(?s)SELECT.*CASE WHEN identity_verified.*FROM "argo_message_attempts".*started_at >= \$1.*started_at <= \$2.*GROUP BY "key".*ORDER BY total DESC`).
		WithArgs(from, to).
		WillReturnRows(sqlmock.NewRows(columns).AddRow("argo-erp", 10, 9, 1, 0, 42.5, to))
	mock.ExpectQuery(`(?s)SELECT.*COALESCE\(CAST\(instance_id AS TEXT\).*FROM "argo_message_attempts".*started_at >= \$1.*started_at <= \$2.*GROUP BY "instance_id".*ORDER BY total DESC`).
		WithArgs(from, to).
		WillReturnRows(sqlmock.NewRows(columns).AddRow("instance-1", 10, 9, 1, 0, 42.5, to))

	applications, instances, err := NewRepository(db).GatewayUsage(context.Background(), argo_model.AttemptFilters{From: from, To: to})
	if err != nil {
		t.Fatal(err)
	}
	if len(applications) != 1 || applications[0].Key != "argo-erp" || len(instances) != 1 || instances[0].Key != "instance-1" {
		t.Fatalf("unexpected usage: applications=%#v instances=%#v", applications, instances)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
