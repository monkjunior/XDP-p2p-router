package db_sqlite

import (
	"fmt"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/vu-ngoc-son/XDP-p2p-router/database"
)

type SQLiteDB struct {
	DB *gorm.DB
}

func NewSQLite(filePath string) (*SQLiteDB, error) {
	err := os.Remove(filePath)
	if err != nil {
		fmt.Printf("error while delete old db %v\n", err)
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(filePath), &gorm.Config{})
	if err != nil {
		fmt.Printf("error while connecting to SQLiteDB db %v\n", err)
		return nil, err
	}
	if err = db.AutoMigrate(database.Hosts{}, database.Peers{}, database.Limits{}); err != nil {
		fmt.Printf("error while migrating SQLiteDB db %v\n", err)
		return nil, err
	}
	return &SQLiteDB{
		DB: db,
	}, nil
}
