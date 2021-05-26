package db_sqlite

import "github.com/vu-ngoc-son/XDP-p2p-router/database"

func (s *SQLiteDB) UpdatePeerLimit(l *database.Limits) error {
	r := s.DB.Model(&l).Where("ip = ?", l.Ip).Updates(&l)
	if r.RowsAffected == 0 {
		r = s.DB.Create(&l)
	}

	return r.Error
}
