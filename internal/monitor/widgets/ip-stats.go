package widgets

import (
	"time"

	"github.com/gizak/termui/v3/widgets"
)

type IPStats struct {
	*widgets.Table
	//DB             *dbSqlite.SQLiteDB
	updateInterval time.Duration
}

func NewIPStats(updateInterval time.Duration) *IPStats {
	self := &IPStats{
		Table:          widgets.NewTable(),
		//DB:             db,
		updateInterval: updateInterval,
	}

	self.Title = "IP Stats"
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

func (s *IPStats) update() {
	s.Rows = [][]string{
		{"ip", "limit"},
		{"ip", "limit"},
		{"ip", "limit"},
		{"ip", "limit"},
		{"ip", "limit"},
	}
	//listIPs, err := s.DB.ListIPsFromLimitsTable(6)
	//if err != nil {
	//
	//	return
	//}
	//s.Rows = append(s.Rows, listIPs...)
}
