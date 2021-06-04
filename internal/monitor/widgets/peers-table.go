package widgets

import (
	"strconv"
	"time"

	goRand "github.com/Pallinder/go-randomdata"
	"github.com/gizak/termui/v3/widgets"
	dbSqlite "github.com/vu-ngoc-son/XDP-p2p-router/database/db-sqlite"
)

type PeersTable struct {
	*widgets.Table
	DB             *dbSqlite.SQLiteDB
	updateInterval time.Duration
}

func NewPeersTable(updateInterval time.Duration, db *dbSqlite.SQLiteDB, fakeData bool) *PeersTable {
	self := &PeersTable{
		Table:          widgets.NewTable(),
		DB:             db,
		updateInterval: updateInterval,
	}

	self.Title = "Peers Stats"
	self.PaddingTop = 1

	self.updatePeersStats(fakeData)
	go func() {
		for range time.NewTicker(self.updateInterval).C {
			self.updatePeersStats(fakeData)
		}
	}()

	return self
}

func (s *PeersTable) updatePeersStats(fakeData bool) {
	s.Rows = [][]string{
		{"Country", "Number of peers"},
	}
	if fakeData {
		s.Rows = append(s.Rows, randomPeerTableData()...)
	}
	//listIPs, err := s.DB.ListIPsFromLimitsTable(6)
	//if err != nil {
	//
	//	return
	//}
	//s.Rows = append(s.Rows, listIPs...)
}

func randomPeerTableData() [][]string {
	labels := []string{"VN", "HK", "US", "UK"}
	data := make([][]string, len(labels))

	for i, l := range labels {
		data[i] = []string{
			l,
			strconv.Itoa(goRand.Number(1000, 20000)),
		}
	}

	return data
}
