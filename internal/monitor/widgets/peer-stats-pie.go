package widgets

import (
	"fmt"
	"math"
	"time"

	goRand "github.com/Pallinder/go-randomdata"
	"github.com/gizak/termui/v3/widgets"
	dbSqlite "github.com/vu-ngoc-son/XDP-p2p-router/database/db-sqlite"
)

const (
	MaxPieParts = 7
)

type PeersPie struct {
	*widgets.PieChart
	Labels         []string
	DB             *dbSqlite.SQLiteDB
	updateInterval time.Duration
	TotalBytes     uint64
}

func NewPeersPie(updateInterval time.Duration, db *dbSqlite.SQLiteDB, fakeData bool) *PeersPie {
	self := &PeersPie{
		PieChart:       widgets.NewPieChart(),
		DB:             db,
		updateInterval: updateInterval,
	}

	self.Title = "Peer Stats"
	self.BorderRight = false
	self.AngleOffset = -0.5 * math.Pi // Where should we start drawing pie
	self.PaddingTop = 1
	self.PaddingRight = 1
	self.PaddingBottom = 1
	self.PaddingLeft = 1

	self.updatePieData(fakeData)
	go func() {
		for range time.NewTicker(self.updateInterval).C {
			self.updatePieData(fakeData)
		}
	}()

	return self
}

func (s *PeersPie) updatePieData(fakeData bool) {
	s.LabelFormatter = func(i int, v float64) string {
		return fmt.Sprintf("%s", s.Labels[i])
	}
	if fakeData {
		s.randomPieData()
		return
	}
	s.crawlPeersPieData()
}

func (s *PeersPie) crawlPeersPieData() {
	countryStats, err := s.DB.ListCountryCodes()
	if err != nil || len(countryStats) <= 0 {
		return
	}

	var pieParts = MaxPieParts
	if MaxPieParts > len(countryStats) {
		pieParts = len(countryStats)
	}

	var totalData uint64 // Data displayed on pie chart

	labels := make([]string, pieParts)
	data := make([]float64, pieParts)
	s.TotalBytes = 0
	for i := 0; i < len(countryStats); i++ {
		s.TotalBytes += countryStats[i].Bytes
	}
	for j := 0; j < pieParts-1; j++ {
		totalData += countryStats[j].Bytes
		labels[j] = countryStats[j].CountryCode
		data[j] = float64(countryStats[j].Bytes) / float64(s.TotalBytes)
	}
	labels[pieParts-1] = "..."
	data[pieParts-1] = float64(s.TotalBytes-totalData) / float64(s.TotalBytes)

	s.Labels = labels
	s.Data = data
}

func (s *PeersPie) randomPieData() {
	s.Labels = []string{"VN", "HK", "US", "UK"}
	s.Data = []float64{
		goRand.Decimal(50, 60),
		goRand.Decimal(15, 20),
		goRand.Decimal(5, 20),
		goRand.Decimal(5, 20),
	}
}
