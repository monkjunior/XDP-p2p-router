package db_sqlite

import (
	"errors"
	"log"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/vu-ngoc-son/XDP-p2p-router/database"
	internalLogger "github.com/vu-ngoc-son/XDP-p2p-router/internal/logger"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	logPath = "log/sqlite.log"
)

type SQLiteDB struct {
	DB       *gorm.DB
	HostInfo *database.Hosts
}

func NewSQLite(filePath string) (*SQLiteDB, error) {
	myLogger := internalLogger.GetLogger()
	err := os.Remove(filePath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			myLogger.Error("error while delete old db", zap.Error(err))
			return nil, err
		}
	}
	err = os.Remove(logPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			myLogger.Error("error while delete log db", zap.Error(err))
			return nil, err
		}
	}

	// TODO: this should be configurable
	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v\n", err)
	}
	gormLogger := logger.New(
		log.New(f, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Duration(500) * time.Millisecond, // Slow SQL threshold
			LogLevel:                  logger.Warn,                           // Log level
			IgnoreRecordNotFoundError: true,                                  // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,                                 // Disable color
		},
	)
	db, err := gorm.Open(sqlite.Open(filePath), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		myLogger.Error("error while connecting to SQLiteDB db %v\n", zap.Error(err))
		return nil, err
	}
	if err = db.AutoMigrate(database.Hosts{}, database.Peers{}, database.Limits{}); err != nil {
		myLogger.Error("error while migrating SQLiteDB db %v\n", zap.Error(err))
		return nil, err
	}
	return &SQLiteDB{
		DB: db,
	}, nil
}
