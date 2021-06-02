package widgets

import (
	"strconv"
	"time"

	goRand "github.com/Pallinder/go-randomdata"
	"github.com/gizak/termui/v3/widgets"
	dbSqlite "github.com/vu-ngoc-son/XDP-p2p-router/database/db-sqlite"
)

type IPStats struct {
	*widgets.Table
	DB             *dbSqlite.SQLiteDB
	updateInterval time.Duration
}

func NewIPStats(updateInterval time.Duration, db *dbSqlite.SQLiteDB, fakeData bool) *IPStats {
	self := &IPStats{
		Table:          widgets.NewTable(),
		DB:             db,
		updateInterval: updateInterval,
	}

	self.Title = "IP Stats"
	self.PaddingTop = 1
	self.PaddingRight = 2

	self.updateIPStats(fakeData)
	go func() {
		for range time.NewTicker(self.updateInterval).C {
			self.updateIPStats(fakeData)
		}
	}()

	return self
}

func (s *IPStats) updateIPStats(fakeData bool) {
	s.Rows = [][]string{
		{"ipv4", "country code", "throughput", "threshold band"},
	}
	if fakeData {
		s.Rows = append(s.Rows, randomIPData(5, 10)...)
	}
	//listIPs, err := s.DB.ListIPsFromLimitsTable(6)
	//if err != nil {
	//
	//	return
	//}
	//s.Rows = append(s.Rows, listIPs...)
}

func randomIPData(minRows, maxRows int) [][]string {
	if maxRows < 0 || minRows > maxRows {
		return nil
	}

	nRows := goRand.Number(minRows, maxRows)
	data := make([][]string, nRows)

	for i := 0; i < nRows; i++ {
		data[i] = []string{
			goRand.IpV4Address(),
			goRand.Country(goRand.TwoCharCountry),
			strconv.Itoa(goRand.Number(10000, 20000)),
			strconv.Itoa(goRand.Number(10000, 20000)),
		}
	}

	return data
}