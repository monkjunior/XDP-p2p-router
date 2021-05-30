package db_sqlite

import (
	"fmt"
	"github.com/vu-ngoc-son/XDP-p2p-router/database"
)

func (s *SQLiteDB) UpdatePeerLimit(l *database.Limits) error {
	r := s.DB.Model(&l).Where("ip = ?", l.Ip).Updates(&l)
	if r.RowsAffected == 0 {
		r = s.DB.Create(&l)
	}

	return r.Error
}

func (s *SQLiteDB) ListIPsFromLimitsTable() ([][]string, error) {
	var listPeers []database.Limits
	result := s.DB.Model(database.Limits{}).Find(&listPeers)
	if result.Error != nil {
		return nil, result.Error
	}

	IPs := make([][]string, len(listPeers))
	for i, p := range listPeers {
		IPs[i] = []string{p.Ip, fmt.Sprintf("%.2f", p.Bandwidth)}
	}
	return IPs, nil
}
