package limit_band

import (
	"go.uber.org/zap"

	bpf "github.com/iovisor/gobpf/bcc"
	bpfMaps "github.com/vu-ngoc-son/XDP-p2p-router/internal/bpf-maps"
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

type BandwidthLimiter struct {
	Table *bpf.Table
}

func NewLimiter(module *bpf.Module) (*BandwidthLimiter, error) {
	return &BandwidthLimiter{
		Table: bpf.NewTable(module.TableId(bpfMaps.LimitBandMap), module),
	}, nil
}

func Close(device string, module *bpf.Module) {
	myLogger := logger.GetLogger()
	if err := module.RemoveXDP(device); err != nil {
		myLogger.Error("failed to remove XDP", zap.String("device", device), zap.Error(err))
	}
	myLogger.Info("close limit bandwidth module successfully")
}

func (p *BandwidthLimiter) ExportMap() (result []bpfMaps.LimitBandMapItem, err error) {
	myLogger := logger.GetLogger()
	countersTable := p.Table

	result = make([]bpfMaps.LimitBandMapItem, 0)
	for item := countersTable.Iter(); item.Next(); {
		if item.Err() != nil {
			myLogger.Error("failed while iterating through item in bpf map", zap.Error(item.Err()))
			continue
		}

		keyRaw := item.Key()
		valueRaw := item.Leaf()

		mapItem := bpfMaps.LimitBandMapItem{
			Key:   bpf.GetHostByteOrder().Uint32(keyRaw),
			Value: bpf.GetHostByteOrder().Uint32(valueRaw),
		}
		result = append(result, mapItem)

	}
	myLogger.Info("export whitelist map successfully", zap.Int("map_length", len(result)))
	return result, nil
}
