package argo_repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestGetMessageMediaUsesInstanceAndProviderMessageID(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(`(?s)SELECT \* FROM "argo_message_media" WHERE instance_id = \$1 AND provider_message_id = \$2 LIMIT \$3`).
		WithArgs("instance-1", "message-1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "instance_id", "provider_message_id", "file_name", "mime_type", "size_bytes", "content"}).
			AddRow("media-1", "instance-1", "message-1", "boleto.pdf", "application/pdf", 4, []byte("%PDF")))

	media, err := NewRepository(db).GetMessageMedia(context.Background(), "instance-1", "message-1")
	if err != nil {
		t.Fatal(err)
	}
	if media == nil || media.FileName != "boleto.pdf" || string(media.Content) != "%PDF" {
		t.Fatalf("unexpected media: %#v", media)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
