package widgets

import (
	"time"

	goRand "github.com/Pallinder/go-randomdata"
	"github.com/gizak/termui/v3/widgets"
	dbSqlite "github.com/vu-ngoc-son/XDP-p2p-router/database/db-sqlite"
)

type WhiteList struct {
	*widgets.Table
	DB             *dbSqlite.SQLiteDB
	updateInterval time.Duration
}

func NewWhiteList(updateInterval time.Duration, db *dbSqlite.SQLiteDB, fakeData bool) *WhiteList {
	self := &WhiteList{
		Table:          widgets.NewTable(),
		DB:             db,
		updateInterval: updateInterval,
	}

	self.Title = "Whitelist"
	self.PaddingTop = 1
	self.PaddingRight = 2

	self.updateWhiteList(fakeData)
	go func() {
		for range time.NewTicker(self.updateInterval).C {
			self.updateWhiteList(fakeData)
		}
	}()

	return self
}

func (s *WhiteList) updateWhiteList(fakeData bool) {
	s.Rows = [][]string{
		{"peer", "state"},
	}
	if fakeData {
		s.Rows = append(s.Rows, randomWhiteListData(5, 10)...)
	}
	//listIPs, err := s.DB.ListIPsFromLimitsTable(6)
	//if err != nil {
	//
	//	return
	//}
	//s.Rows = append(s.Rows, listIPs...)
}

func randomWhiteListData(minRows, maxRows int) [][]string {
	if maxRows < 0 || minRows > maxRows {
		return nil
	}

	nRows := goRand.Number(minRows, maxRows)
	data := make([][]string, nRows)

	for i := 0; i < nRows; i++ {
		state := "XDP PASS"
		block := goRand.Boolean()
		if block {
			state = "XDP DROP"
		}
		data[i] = []string{
			goRand.IpV4Address(),
			state,
		}
	}

	return data
}
