package postgres_test_utils

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostgresTestRow struct {
	ID        uint64    `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"index"`
	Size      int
	Data      []byte `gorm:"type:bytea"`
}

func Connect(uri string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(uri), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&PostgresTestRow{}); err != nil {
		return nil, err
	}
	return db, nil
}
