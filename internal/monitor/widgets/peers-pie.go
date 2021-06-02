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
	self.update(fakeData)
	go func() {
		for range time.NewTicker(self.updateInterval).C {
			self.update(fakeData)
		}
	}()

	return self
}

func (s *PeersPie) update(fakeData bool) {
	if fakeData {
		s.Data, s.Labels = randomPieData()
		s.LabelFormatter = func(i int, v float64) string {
			return fmt.Sprintf("%s", s.Labels[i])
		}
	}
	//listIPs, err := s.DB.ListIPsFromLimitsTable(6)
	//if err != nil {
	//
	//	return
	//}
	//s.Rows = append(s.Rows, listIPs...)
}

func randomPieData() ([]float64, []string) {
	labels := []string{"VN", "HK", "US", "UK"}
	data := []float64{
		goRand.Decimal(50, 60),
		goRand.Decimal(15, 20),
		goRand.Decimal(5, 20),
		goRand.Decimal(5, 20),
	}
	return data, labels
}
