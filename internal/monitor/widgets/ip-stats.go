package widgets

import (
	"fmt"
	bpfMaps "github.com/vu-ngoc-son/XDP-p2p-router/internal/bpf-maps"
	"strconv"
	"sync"
	"time"

	goRand "github.com/Pallinder/go-randomdata"
	"github.com/gizak/termui/v3/widgets"
	bpf "github.com/iovisor/gobpf/bcc"
	dbSqlite "github.com/vu-ngoc-son/XDP-p2p-router/database/db-sqlite"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/common"
)

type IPStats struct {
	*widgets.Table
	DB             *dbSqlite.SQLiteDB
	PktCaptureMap  *bpf.Table
	IPWhitelistMap *bpf.Table
	UpdateInterval time.Duration
	throughputMap  *sync.Map

	IPsPool map[uint32]chan bpfMaps.PktCounterValue
}

func NewIPStats(t time.Duration, db *dbSqlite.SQLiteDB, pktCap, whitelist *bpf.Table, fakeData bool) *IPStats {
	self := &IPStats{
		Table:          widgets.NewTable(),
		DB:             db,
		PktCaptureMap:  pktCap,
		IPWhitelistMap: whitelist,
		UpdateInterval: t,
		throughputMap:  &sync.Map{},
	}

	self.Title = "IP Stats"
	self.PaddingTop = 1
	self.PaddingRight = 2

	self.updateIPStats(fakeData)
	go func() {
		for range time.NewTicker(self.UpdateInterval).C {
			self.updateIPStats(fakeData)
		}
	}()

	return self
}

func (s *IPStats) updateIPStats(fakeData bool) {
	s.Rows = [][]string{
		{"IPv4", "Country Code", "Distance (kms)", "Throughput (bps)", "Bandwidth"},
	}
	if fakeData {
		s.randomIPData(5, 10)
		return
	}

	s.crawlIPData()
}

func (s *IPStats) crawlIPData() {
	wg := sync.WaitGroup{}
	for item := s.IPWhitelistMap.Iter(); item.Next(); {
		if item.Err() != nil {
			continue
		}
		wg.Add(1)
		go func(group *sync.WaitGroup) {
			defer group.Done()

			peerIP, err := common.ConvertUint8ToUInt32(item.Key())
			if err != nil {
				return
			}

			_, exist := s.throughputMap.Load(peerIP)
			if exist {
				return
			}

			go func(k []byte, key uint32, m *sync.Map) {
				prev := 0.0
				clock := time.NewTicker(time.Second)
				interval := 1.0
				for {
					select {
					case <-clock.C:
						counterData, err := s.PktCaptureMap.Get(k)
						if err != nil {
							interval += 1
							continue
						}
						current, err := common.ConvertUint8ToUInt64(counterData[8:16])
						if err != nil {
							interval += 1
							continue
						}
						m.Store(key, (float64(current)-prev)*8/interval)
						prev = float64(current)
						interval = 1.0
					}
				}
			}(item.Key(), peerIP, s.throughputMap)
		}(&wg)
	}
	wg.Wait()

	s.throughputMap.Range(func(k, v interface{}) bool {
		p, err := s.DB.GetPeer(k.(uint32))
		if err != nil {
			return true
		}
		s.Rows = append(s.Rows, []string{
			fmt.Sprintf("%s", p.IpAddress),
			p.CountryCode,
			fmt.Sprintf("%.2f", p.Distance),
			fmt.Sprintf("%.2f", v),
			fmt.Sprintf("%.2f", p.Bandwidth),
		})
		return true
	})
	return
}

func (s *IPStats) randomIPData(minRows, maxRows int) {
	if maxRows < 0 || minRows > maxRows {
		return
	}

	nRows := goRand.Number(minRows, maxRows)

	for i := 0; i < nRows; i++ {
		s.Rows = append(s.Rows, []string{
			goRand.IpV4Address(),
			goRand.Country(goRand.TwoCharCountry),
			strconv.Itoa(goRand.Number(10000, 20000)),
			strconv.Itoa(goRand.Number(10000, 20000)),
			strconv.Itoa(goRand.Number(10000, 20000)),
		})
	}

	return
}
