package ip2location

import (
	"fmt"
	dbSqlite "github.com/vu-ngoc-son/XDP-p2p-router/database/db-sqlite"
	"github.com/vu-ngoc-son/XDP-p2p-router/database/geolite2"
	packetCapture "github.com/vu-ngoc-son/XDP-p2p-router/internal/packet-capture"
	"sync"
)

type Locator struct {
	PacketCapture *packetCapture.PacketCapture
	DB            *dbSqlite.SQLiteDB
	GeoDB         *geolite2.GeoLite2
}

func NewLocator(p *packetCapture.PacketCapture, db *dbSqlite.SQLiteDB, g *geolite2.GeoLite2) *Locator {
	return &Locator{
		PacketCapture: p,
		DB:            db,
		GeoDB:         g,
	}
}

func (l *Locator) UpdatePeersToDB() {
	pktCounterMap, err := l.PacketCapture.ExportMap()
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(pktCounterMap))
	for _, item := range pktCounterMap {
		peer, err := l.GeoDB.IPInfo(item.Key.SourceAddr)
		if err != nil {
			return
		}
		go l.DB.UpdateOrCreatePeer(peer, &wg)
	}
	wg.Wait()
	fmt.Println("update peers from bpf map to sqlite db successfully")
	return
}
