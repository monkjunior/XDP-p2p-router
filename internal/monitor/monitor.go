package monitor

import (
	"github.com/iovisor/gobpf/bcc"
	dbSqlite "github.com/vu-ngoc-son/XDP-p2p-router/database/db-sqlite"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/common"
	limitBand "github.com/vu-ngoc-son/XDP-p2p-router/internal/limit-band"
	packetCapture "github.com/vu-ngoc-son/XDP-p2p-router/internal/packet-capture"
)

type PeerStat struct {
	MapKey     []byte
	ThroughPut uint64
}

type Monitor struct {
	PacketCapture *packetCapture.PacketCapture
	LimitBand     *limitBand.BandwidthLimiter
	DB            *dbSqlite.SQLiteDB
	AddrPool      map[string]PeerStat
}

func NewMonitor(p *packetCapture.PacketCapture, l *limitBand.BandwidthLimiter, db *dbSqlite.SQLiteDB) *Monitor {
	addrPool := make(map[string]PeerStat)
	return &Monitor{
		PacketCapture: p,
		LimitBand:     l,
		DB:            db,
		AddrPool:      addrPool,
	}
}

func (m *Monitor) updatePool() error {
	listIPs, err := m.DB.ListIPsFromLimitsTable(5)
  
	if err != nil {
		return err
	}

	for _, ip := range listIPs {
		if _, exist := m.AddrPool[ip[0]]; !exist {
			ipUint32, err := common.ConvertIPToUint32(ip[0])
			if err != nil {
				continue
			}
			mapKey := make([]byte, 4)
			bcc.GetHostByteOrder().PutUint32(mapKey, ipUint32)
			m.AddrPool[ip[0]] = PeerStat{
				MapKey:     mapKey,
				ThroughPut: 0,
			}
		}
	}

	return nil
}

func (m *Monitor) IPList() [][]string {
	throughputTable := [][]string{
		{"ip"},
	}

	err := m.updatePool()
	if err != nil {
		return throughputTable
	}

	for k := range m.AddrPool {
		throughputTable = append(throughputTable, []string{
			k,
		})
	}

	return throughputTable
}
