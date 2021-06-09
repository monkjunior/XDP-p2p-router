package widgets

import (
	"fmt"
	"sync"
	"time"

	goRand "github.com/Pallinder/go-randomdata"
	"github.com/gizak/termui/v3/widgets"
	bpf "github.com/iovisor/gobpf/bcc"
	dbSqlite "github.com/vu-ngoc-son/XDP-p2p-router/database/db-sqlite"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/common"
)

type WhiteList struct {
	*widgets.Table
	PktCaptureMap  *bpf.Table
	IPWhitelistMap *bpf.Table
	DB             *dbSqlite.SQLiteDB
	updateInterval time.Duration
	whitelistMap   *sync.Map
}

func NewWhiteList(updateInterval time.Duration, db *dbSqlite.SQLiteDB, pktCap, whitelist *bpf.Table, fakeData bool) *WhiteList {
	self := &WhiteList{
		Table:          widgets.NewTable(),
		DB:             db,
		PktCaptureMap:  pktCap,
		IPWhitelistMap: whitelist,
		updateInterval: updateInterval,
		whitelistMap:   &sync.Map{},
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
		{"Peer", "State"},
	}
	if fakeData {
		s.Rows = append(s.Rows, randomWhiteListData(5, 10)...)
		return
	}

	s.crawlWhitelistData()
}

func (s *WhiteList) crawlWhitelistData() {
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

			_, exist := s.whitelistMap.Load(peerIP)
			if exist {
				return
			}

			go func(k []byte, key uint32, m *sync.Map) {
				clock := time.NewTicker(time.Second)
				peerStateRaw, err := s.IPWhitelistMap.Get(k)
				if err != nil {
					return
				}
				peerStateInt, err := common.ConvertUint8ToUInt32(peerStateRaw)
				if err != nil {
					return
				}
				peerIPv4, err := common.ConvertUint8ToIP(k)
				if err != nil {
					return
				}
				for {
					select {
					case <-clock.C:
						switch peerStateInt {
						case 0:
							m.Store(peerIPv4, "XDP_ABORTED")
						case 1:
							m.Store(peerIPv4, "XDP_DROP")
						case 2:
							m.Store(peerIPv4, "XDP_PASS")
						case 3:
							m.Store(peerIPv4, "XDP_TX")
						case 4:
							m.Store(peerIPv4, "XDP_REDIRECT")
						default:
							m.Store(peerIPv4, peerStateInt)
						}
					}
				}
			}(item.Key(), peerIP, s.whitelistMap)
		}(&wg)
	}
	wg.Wait()

	s.whitelistMap.Range(func(k, v interface{}) bool {
		s.Rows = append(s.Rows, []string{
			fmt.Sprintf("%s", k),
			fmt.Sprintf("%s", v),
		})
		return true
	})
	return
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
