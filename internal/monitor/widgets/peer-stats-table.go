package widgets

import (
	"fmt"
	"time"

	"github.com/gizak/termui/v3/widgets"
)

type PeersTable struct {
	*PeersPie
	*widgets.Table
}

func NewPeersTable(pie *PeersPie) *PeersTable {
	self := &PeersTable{
		PeersPie: pie,
		Table:    widgets.NewTable(),
	}

	self.BorderLeft = false
	self.PaddingTop = 1

	self.updatePeersStats()
	go func() {
		for range time.NewTicker(self.updateInterval).C {
			self.updatePeersStats()
		}
	}()

	return self
}

func (s *PeersTable) updatePeersStats() {
	s.Rows = [][]string{
		{"Country", "Data percent (%)"},
	}

	for i := 0; i < len(s.Labels); i++ {
		s.Rows = append(s.Rows, []string{
			fmt.Sprintf("%s", s.PeersPie.Labels[i]),
			fmt.Sprintf("%.2f", s.PeersPie.Data[i]*100),
		})
	}
}
