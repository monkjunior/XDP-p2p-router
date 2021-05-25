package db_sqlite

import (
	"github.com/vu-ngoc-son/XDP-p2p-router/database"
	"sync"
)

func (s *SQLiteDB) UpdateOrCreatePeer(peer *database.Peers, wg *sync.WaitGroup) {
	defer wg.Done()

	r := s.DB.Model(peer).Where("ip = ?", peer.Ip).Updates(peer)
	if r.RowsAffected == 0 {
		r = s.DB.Model(database.Peers{}).Create(peer)
	}
}

func (s *SQLiteDB) AddPeers(peers []database.Peers) error {
	r := s.DB.Model(database.Peers{}).CreateInBatches(peers, 20)
	return r.Error
}
