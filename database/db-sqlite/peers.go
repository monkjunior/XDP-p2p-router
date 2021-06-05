package db_sqlite

import (
	"math"
	"sync"

	"github.com/vu-ngoc-son/XDP-p2p-router/database"
)

func (s *SQLiteDB) UpdateOrCreatePeer(peer *database.Peers, wg *sync.WaitGroup) {
	defer wg.Done()

	var prev []database.Peers
	r := s.DB.Model(peer).Where("ip_number = ?", peer.IpNumber).First(&prev)
	if r.RowsAffected == 0 {
		r = s.DB.Model(database.Peers{}).Create(peer)
		return
	}
	if len(prev) == 1 {
		r = s.DB.Model(database.Peers{}).Where("ip_number = ?", peer.IpNumber).Updates(peer)
		return
	}
}

func (s *SQLiteDB) GetPeer(IP uint32) (database.Peers, error) {
	var peer []database.Peers
	result := s.DB.Model(database.Peers{}).Where("ip_number = ?", IP).First(&peer)
	if result.Error != nil {
		return database.Peers{}, result.Error
	}

	return peer[0], nil
}

func (s *SQLiteDB) GetPeers() ([]database.Peers, error) {
	var peers []database.Peers
	result := s.DB.Model(database.Peers{}).Find(&peers)
	if result.Error != nil {
		return nil, result.Error
	}

	return peers, nil
}

func (s *SQLiteDB) AddPeers(peers []database.Peers) error {
	r := s.DB.Model(database.Peers{}).CreateInBatches(peers, 20)
	return r.Error
}

func (s *SQLiteDB) FindNearByPeers() (n1, n2, n3 float64, err error) {
	var host database.Hosts
	r := s.DB.Model(database.Hosts{}).First(&host)
	if r.Error != nil {
		return math.NaN(), math.NaN(), math.NaN(), err
	}
	var r1, r2, r3 int64
	s.DB.Table("peers").Where("asn = ?", host.Asn).Count(&r1)
	s.DB.Table("peers").Where("isp = ?", host.Isp).Count(&r2)
	s.DB.Table("peers").Where("country_code = ?", host.CountryCode).Count(&r3)
	return float64(r1), float64(r2), float64(r3), nil
}

func (s *SQLiteDB) CompareToHost(peerInfo database.Peers) (sameASN, sameISP, sameCountry bool) {
	var host database.Hosts
	r := s.DB.Model(database.Hosts{}).First(&host)
	if r.Error != nil {
		return false, false, false
	}
	sameASN = peerInfo.Asn == host.Asn
	sameISP = peerInfo.Isp == host.Isp
	sameCountry = peerInfo.CountryCode == host.CountryCode
	return sameASN, sameISP, sameCountry
}
