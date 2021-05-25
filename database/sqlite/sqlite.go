package sqlite

import (
	"fmt"
	"github.com/vu-ngoc-son/XDP-p2p-router/database"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SQLite struct {
	DB *gorm.DB
}

func NewSQLite(filePath string) (*SQLite, error) {
	db, err := gorm.Open(sqlite.Open(filePath), &gorm.Config{})
	if err != nil {
		fmt.Printf("error while connecting to SQLite db %v", err)
		return nil, err
	}
	if err = db.AutoMigrate(database.Hosts{}, database.Peers{}, database.Limits{}); err != nil {
		fmt.Printf("error while migrating SQLite db %v", err)
		return nil, err
	}
	return &SQLite{
		DB: db,
	}, nil
}
