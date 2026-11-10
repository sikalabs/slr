package postgres_test_utils

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Message struct {
	ID        uint64    `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"index"`
	Prefix    string
	Message   string
}

type Counter struct {
	ID        uint64    `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"index"`
	Prefix    string
	Counter   uint64
}

func (Counter) TableName() string {
	return "counter"
}

type PostgresTestRow struct {
	ID        uint64    `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"index"`
	Size      int
	Data      []byte `gorm:"type:bytea"`
}

func Connect(uri string, models ...interface{}) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(uri), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(models...); err != nil {
		return nil, err
	}
	return db, nil
}
