package db_sqlite

import (
	"github.com/vu-ngoc-son/XDP-p2p-router/database"
)

func (s *SQLiteDB) CreateHost(host *database.Hosts) error {
	r := s.DB.Model(database.Hosts{}).Create(host)
	return r.Error
}
