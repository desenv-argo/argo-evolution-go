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

func TestHeartbeatMetricsMapsPostgresAggregateTypes(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("open sqlmock db: %v", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open gorm db: %v", err)
	}

	mock.ExpectQuery(`(?s)SELECT.*COUNT\(\*\) AS heartbeat_events.*DOUBLE PRECISION.*FROM "argo_integration_heartbeats".*received_at >= \$1.*received_at <= \$2`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"heartbeat_events", "unhealthy_events", "average_latency_ms"}).
			AddRow(int64(7), int64(2), float64(125.5)))

	from := time.Date(2026, time.August, 11, 20, 29, 0, 0, time.UTC)
	to := from.Add(7 * 24 * time.Hour)
	events, unhealthy, average, err := NewRepository(gormDB).HeartbeatMetrics(context.Background(), argo_model.HeartbeatFilters{From: from, To: to})
	if err != nil {
		t.Fatalf("heartbeat metrics: %v", err)
	}
	if events != 7 || unhealthy != 2 || average != 125.5 {
		t.Fatalf("unexpected metrics: events=%d unhealthy=%d average=%f", events, unhealthy, average)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
