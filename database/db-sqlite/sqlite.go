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
	runMode = "file"
	logPath = "log/sqlite.log"

	dbPath string

	gormLogger *log.Logger
)

type SQLiteDB struct {
	DB       *gorm.DB
	HostInfo *database.Hosts
}

func initSQLite() {
	myLogger := internalLogger.GetLogger()

	switch runMode {
	case "file":
		dbPath = "data/sqlite/p2p-router.db"
	case "memory":
		dbPath = "file:p2p-router.db?mode=memory&cache=shared"
	}

	err := os.Remove(dbPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			myLogger.Error("error while delete old db", zap.Error(err))
			return
		}
	}
	err = os.Remove(logPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			myLogger.Error("error while delete log db", zap.Error(err))
			return
		}
	}

	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		myLogger.Fatal("error opening file", zap.Error(err))
		return
	}
	gormLogger = log.New(f, "\r\n", log.LstdFlags) // io writer
}

func NewSQLite() (*SQLiteDB, error) {
	initSQLite()

	myLogger := internalLogger.GetLogger()

	gormLogger := logger.New(
		gormLogger,
		logger.Config{
			SlowThreshold:             time.Duration(500) * time.Millisecond, // Slow SQL threshold
			LogLevel:                  logger.Warn,                           // Log level
			IgnoreRecordNotFoundError: true,                                  // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,                                 // Disable color
		},
	)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
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
