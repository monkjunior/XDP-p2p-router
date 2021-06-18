package packet_capture

import (
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	bpf "github.com/iovisor/gobpf/bcc"
	dbSqlite "github.com/vu-ngoc-son/XDP-p2p-router/database/db-sqlite"
	"github.com/vu-ngoc-son/XDP-p2p-router/database/geolite2"
	bpfMaps "github.com/vu-ngoc-son/XDP-p2p-router/internal/bpf-maps"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/common"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/logger"
)

/*
#ifdef asm_inline
#undef asm_inline
#define asm_inline asm
#endif
#cgo CFLAGS: -I/usr/include/bcc/compat
#cgo LDFLAGS: -lbcc
#include <bcc/bcc_common.h>
#include <bcc/libbpf.h>
*/
import "C"

const (
	UpdateIntervalSeconds = 5
)

type PacketCapture struct {
	Device  string
	Module  *bpf.Module
	Table   *bpf.Table
	IPsPool map[uint32]bpfMaps.PktCounterValue
	DB      *dbSqlite.SQLiteDB
	GeoDB   *geolite2.GeoLite2
}

func Start(device string, module *bpf.Module, db *dbSqlite.SQLiteDB, g *geolite2.GeoLite2) (*PacketCapture, error) {
	myLogger := logger.GetLogger()
	fn, err := module.Load("packet_counter", C.BPF_PROG_TYPE_XDP, 1, 65536)
	if err != nil {
		myLogger.Error("failed to load xdp program packetCounter: ", zap.Error(err))
		return nil, err
	}

	err = module.AttachXDP(device, fn)
	if err != nil {
		myLogger.Error("failed to attach xdp program: ", zap.Error(err))
		return nil, err
	}

	self := &PacketCapture{
		Device:  device,
		Module:  module,
		Table:   bpf.NewTable(module.TableId(bpfMaps.PacketCaptureMap), module),
		IPsPool: make(map[uint32]bpfMaps.PktCounterValue),
		DB:      db,
		GeoDB:   g,
	}

	self.updateIPsPool()

	go func() {
		for range time.NewTicker(UpdateIntervalSeconds * time.Second).C {
			self.updateIPsPool()
		}
	}()

	return self, nil
}

func (p *PacketCapture) updateIPsPool() {
	myLogger := logger.GetLogger()
	countersTable := p.Table

	for item := countersTable.Iter(); item.Next(); {
		if item.Err() != nil {
			myLogger.Error("failed while iterating through item in bpf map", zap.Error(item.Err()))
			continue
		}

		keyRaw := item.Key()

		key, err := common.ConvertUint8ToUInt32(keyRaw)
		if err != nil {
			myLogger.Error("failed while converting key to uint32", zap.Error(err))
			continue
		}

		value, err := p.getValueFromKey(keyRaw)
		if err != nil {
			myLogger.Error("failed to get key from bpf map",
				zap.Binary("key", keyRaw),
				zap.Error(err),
			)
			continue
		}
		p.IPsPool[key] = *value
	}

	err := p.updatePeersDB()
	if err != nil {
		myLogger.Error("failed update peers db",
			zap.Error(err),
		)
	}
	return
}

func (p *PacketCapture) getValueFromKey(keyRaw []uint8) (*bpfMaps.PktCounterValue, error) {
	myLogger := logger.GetLogger()
	countersTable := p.Table
	valueRaw, err := countersTable.Get(keyRaw)
	if err != nil {
		myLogger.Error("failed to get key from bpf map",
			zap.Binary("key", keyRaw),
			zap.Error(err),
		)
		return nil, err
	}

	rxPackets, err := common.ConvertUint8ToUInt64(valueRaw[0:8])
	if err != nil {
		myLogger.Error("failed while converting rxPackets to uint64",
			zap.Binary("key", keyRaw),
			zap.Binary("rx_packets", valueRaw[0:8]),
			zap.Error(err))
		return nil, err
	}

	rxBytes, err := common.ConvertUint8ToUInt64(valueRaw[8:16])
	if err != nil {
		myLogger.Error("failed while converting rxBytes to uint64",
			zap.Binary("key", keyRaw),
			zap.Binary("rx_bytes", valueRaw[8:16]),
			zap.Error(err))
		return nil, err
	}

	return &bpfMaps.PktCounterValue{
		RxPackets: rxPackets,
		RxBytes:   rxBytes,
	}, nil
}

func (p *PacketCapture) updatePeersDB() (err error) {
	myLogger := logger.GetLogger()
	countersTable := p.Table
	var wg sync.WaitGroup
	for item := countersTable.Iter(); item.Next(); {
		if item.Err() != nil {
			myLogger.Error("failed while iterating through item in bpf map", zap.Error(item.Err()))
			continue
		}

		keyRaw := item.Key()
		valueRaw := item.Leaf()

		key, err := common.ConvertUint8ToUInt32(keyRaw)
		if err != nil {
			myLogger.Error("failed while converting key to uint32", zap.Error(err))
			continue
		}

		rxPackets, err := common.ConvertUint8ToUInt64(valueRaw[0:8])
		if err != nil {
			myLogger.Error("failed while converting rxPackets to uint64", zap.Error(err))
			continue
		}

		rxBytes, err := common.ConvertUint8ToUInt64(valueRaw[8:16])
		if err != nil {
			myLogger.Error("failed while converting rxBytes to uint64", zap.Error(err))
			continue
		}

		IPAddress, err := common.ConvertUint8ToIP(keyRaw)
		if err != nil {
			myLogger.Error("failed while parsing key to IPv4", zap.Error(err))
			continue
		}

		IP := net.ParseIP(IPAddress)
		if IP.IsUnspecified() {
			err := countersTable.Delete(keyRaw)
			if err != nil {
				myLogger.Error("failed while delete unspecified IPv4 key", zap.Error(err))
				continue
			}
			myLogger.Info("deleted unspecified IPv4 key", zap.Binary("key", keyRaw))
			continue
		}

		IPNumber := key
		peer, err := p.GeoDB.IPInfo(IP, IPNumber, rxPackets, rxBytes)
		if err != nil {
			myLogger.Error("failed to get peer info", zap.Error(err))
			return err
		}
		wg.Add(1)
		go p.DB.UpdateOrCreatePeer(peer, &wg)
	}
	wg.Wait()
	myLogger.Info("export packet capture map successfully")
	return nil
}

func (p *PacketCapture) Close() {
	myLogger := logger.GetLogger()
	if err := p.Module.RemoveXDP(p.Device); err != nil {
		myLogger.Error("failed to remove XDP", zap.String("device", p.Device), zap.Error(err))
	}
	myLogger.Info("close packet capture module successfully")
}
