package monitor

import (
	"fmt"
	"github.com/iovisor/gobpf/bcc"
	dbSqlite "github.com/vu-ngoc-son/XDP-p2p-router/database/db-sqlite"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/common"
	limitBand "github.com/vu-ngoc-son/XDP-p2p-router/internal/limit-band"
	packetCapture "github.com/vu-ngoc-son/XDP-p2p-router/internal/packet-capture"
	"sync"
	"time"
)

type Monitor struct {
	PacketCapture *packetCapture.PacketCapture
	LimitBand     *limitBand.BandwidthLimiter
	DB            *dbSqlite.SQLiteDB
	AddrPool      []uint32
}

func NewMonitor(p *packetCapture.PacketCapture, l *limitBand.BandwidthLimiter, db *dbSqlite.SQLiteDB) *Monitor {
	addrPool := make([]uint32, 0)
	return &Monitor{
		PacketCapture: p,
		LimitBand:     l,
		DB:            db,
		AddrPool:      addrPool,
	}
}

func (m *Monitor) UpdatePool(interval int, quit *chan bool) {
	go func() {
		for {
			time.Sleep(time.Duration(interval) * time.Second)
			listIPs, err := m.DB.ListIPsFromLimitsTable()
			if err != nil {
				continue
			}

			for _, ip := range listIPs {
				ipUint32, err := common.ConvertIPToUint32(ip)
				if err != nil {
					fmt.Println(err)
					continue
				}
				m.AddrPool = append(m.AddrPool, ipUint32)
			}
		}
	}()

	<-*quit
	fmt.Println("stop updating throughput")
}

func (m *Monitor) ExportThroughput(quitAll *chan int, interval int) {
	quitUpdate := make(chan bool)
	quitCrawl := make(chan bool)
	go m.UpdatePool(interval, &quitUpdate)
	go m.CrawlThroughPut(&quitCrawl, interval)

	<-*quitAll
	quitUpdate <- true
	quitCrawl <- true
	fmt.Println("stop export throughput")
}

func (m *Monitor) CrawlThroughPut(quit *chan bool, interval int) {
	go func() {
		for {
			addrPool := m.AddrPool

			wg := sync.WaitGroup{}
			wg.Add(len(addrPool))
			for _, ipUint32 := range addrPool {
				go func(wg *sync.WaitGroup, ipUint32 uint32) {
					k := make([]byte, 4)
					bcc.GetHostByteOrder().PutUint32(k, ipUint32)
					prevRxBytes, err := m.PacketCapture.Table.Get(k)
					if err != nil {
						fmt.Println("err while getting prev", err)
						return
					}
					time.Sleep(time.Duration(interval) * time.Second)
					currentRxBytes, err := m.PacketCapture.Table.Get(k)
					if err != nil {
						fmt.Println("err while getting current", err)
						return
					}
					prevRx, err := common.ConvertUint8ToUInt64(prevRxBytes[8:16])
					if err != nil {
						fmt.Println("err while converting prev", err)
						return
					}
					currentRx, err := common.ConvertUint8ToUInt64(currentRxBytes[8:16])
					if err != nil {
						fmt.Println("err while converting current", err)
						return
					}
					fmt.Printf("%v \t %d \tbps \n", k, currentRx-prevRx)
					wg.Done()
				}(&wg, ipUint32)
			}

			wg.Wait()
		}
	}()

	<-*quit
	fmt.Println("stop crawling throughput")
}
