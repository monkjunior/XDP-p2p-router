package packet_capture

import (
	"go.uber.org/zap"

	bpf "github.com/iovisor/gobpf/bcc"
	bpfMaps "github.com/vu-ngoc-son/XDP-p2p-router/internal/bpf-maps"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/common"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/logger"
)

/*
#cgo CFLAGS: -I/usr/include/bcc/compat
#cgo LDFLAGS: -lbcc
#include <bcc/bcc_common.h>
#include <bcc/libbpf.h>
*/
import "C"

type PacketCapture struct {
	Table *bpf.Table
}

func Start(device string, module *bpf.Module) (*PacketCapture, error) {
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

	return &PacketCapture{
		Table: bpf.NewTable(module.TableId(bpfMaps.PacketCaptureMap), module),
	}, nil
}

func Close(device string, module *bpf.Module) {
	myLogger := logger.GetLogger()
	if err := module.RemoveXDP(device); err != nil {
		myLogger.Error("failed to remove XDP", zap.String("device", device), zap.Error(err))
	}
	myLogger.Info("close packet capture module successfully")
}

func (p *PacketCapture) ExportMap() (result []bpfMaps.PktCounterMapItem, err error) {
	myLogger := logger.GetLogger()
	countersTable := p.Table

	result = make([]bpfMaps.PktCounterMapItem, 0)
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
			return nil, err
		}

		rxPackets, err := common.ConvertUint8ToUInt64(valueRaw[0:8])
		if err != nil {
			myLogger.Error("failed while converting rxPackets to uint64", zap.Error(err))
			return nil, err
		}

		rxBytes, err := common.ConvertUint8ToUInt64(valueRaw[8:16])
		if err != nil {
			myLogger.Error("failed while converting rxBytes to uint64", zap.Error(err))
			return nil, err
		}

		mapItem := bpfMaps.PktCounterMapItem{
			Key: key,
			Value: bpfMaps.PktCounterValue{
				RxPackets: rxPackets,
				RxBytes:   rxBytes,
			},
		}
		result = append(result, mapItem)

	}
	myLogger.Info("export packet capture map successfully", zap.Int("map_length", len(result)))
	return result, nil
}
