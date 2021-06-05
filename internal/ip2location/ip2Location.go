package ip2location

import (
	"sync"

	"go.uber.org/zap"

	dbSqlite "github.com/vu-ngoc-son/XDP-p2p-router/database/db-sqlite"
	"github.com/vu-ngoc-son/XDP-p2p-router/database/geolite2"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/logger"
	packetCapture "github.com/vu-ngoc-son/XDP-p2p-router/internal/packet-capture"
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
	myLogger := logger.GetLogger()
	pktCounterMap, err := l.PacketCapture.ExportMap()
	if err != nil {
		myLogger.Error("failed to export packet capture bpf map", zap.Error(err))
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(pktCounterMap))
	for _, item := range pktCounterMap {
		peer, err := l.GeoDB.IPInfo(item.Key)
		if err != nil {
			myLogger.Error("failed to get peer info", zap.Error(err))
			return
		}
		go l.DB.UpdateOrCreatePeer(peer, &wg)
	}
	wg.Wait()
	myLogger.Info("update peers to DB successfully", zap.Int("map_length", len(pktCounterMap)))
	return
}
