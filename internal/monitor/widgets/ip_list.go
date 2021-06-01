package widgets

import (
	"time"

	"github.com/gizak/termui/v3/widgets"
	dbSqlite "github.com/vu-ngoc-son/XDP-p2p-router/database/db-sqlite"
)

type IPList struct {
	*widgets.Table
	DB             *dbSqlite.SQLiteDB
	updateInterval time.Duration
}

func NewIPList(db *dbSqlite.SQLiteDB, updateInterval time.Duration) *IPList {
	self := &IPList{
		Table:          widgets.NewTable(),
		DB:             db,
		updateInterval: updateInterval,
	}

	self.Title = "top 5 bandwidths of ips"
	self.ColumnResizer()
	self.SetRect(5, 5, 60, 20)

	self.update()

	go func() {
		for range time.NewTicker(self.updateInterval).C {
			self.update()
		}
	}()

	return self
}

func (s *IPList) update() {
	s.Rows = [][]string{
		{"ip", "limit"},
	}
	listIPs, err := s.DB.ListIPsFromLimitsTable(6)
	if err != nil {

		return
	}
	s.Rows = append(s.Rows, listIPs...)
}
