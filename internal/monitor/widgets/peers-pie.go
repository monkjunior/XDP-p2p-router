package widgets

import (
	"fmt"
	"math"
	"time"

	goRand "github.com/Pallinder/go-randomdata"
	"github.com/gizak/termui/v3/widgets"
	dbSqlite "github.com/vu-ngoc-son/XDP-p2p-router/database/db-sqlite"
)

type PeersPie struct {
	*widgets.PieChart
	Labels         []string
	DB             *dbSqlite.SQLiteDB
	updateInterval time.Duration
}

func NewPeersPie(updateInterval time.Duration, db *dbSqlite.SQLiteDB, fakeData bool) *PeersPie {
	self := &PeersPie{
		PieChart:       widgets.NewPieChart(),
		DB:             db,
		updateInterval: updateInterval,
	}

	self.Title = "Peers Population"
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
		return fmt.Sprintf("%s-%0.2f", s.Labels[i], v)
	}
	if fakeData {
		s.randomPieData()
		return
	}
	s.crawlPeersPieData()
}

func (s *PeersPie) crawlPeersPieData() {
	countryStats, err := s.DB.ListCountryCodes()
	if err != nil {
		return
	}

	var totalBytes uint64

	labels := make([]string, len(countryStats))
	data := make([]float64, len(countryStats))
	for i := 0; i < len(countryStats); i++ {
		labels[i] = countryStats[i].CountryCode
		totalBytes += countryStats[i].Bytes
	}
	for j := 0; j < len(countryStats); j++ {
		data[j] = float64(countryStats[j].Bytes)/float64(totalBytes)
	}
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
