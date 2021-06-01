package db_sqlite

import (
	"errors"
	"fmt"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/vu-ngoc-son/XDP-p2p-router/database"
)

type SQLiteDB struct {
	DB       *gorm.DB
	HostInfo *database.Hosts
}

func NewSQLite(filePath string) (*SQLiteDB, error) {
	err := os.Remove(filePath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Printf("error while delete old db %v\n", err)
			return nil, err
		}
	}

	db, err := gorm.Open(sqlite.Open(filePath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
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
