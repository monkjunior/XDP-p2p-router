package limit_band

import (
	"fmt"
	bpf "github.com/iovisor/gobpf/bcc"
	bpfMaps "github.com/vu-ngoc-son/XDP-p2p-router/internal/bpf-maps"
	"os"
)

/*
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
	if err := module.RemoveXDP(device); err != nil {
		fmt.Fprintf(os.Stderr, "failed to remove XDP from %s: %v\n", device, err)
	}
	fmt.Println("close limit bandwidth module successfully")
}

func (p *BandwidthLimiter) ExportMap() (result []bpfMaps.LimitBandMapItem, err error) {
	countersTable := p.Table

	result = make([]bpfMaps.LimitBandMapItem, 0)
	for item := countersTable.Iter(); item.Next(); {
		keyRaw := item.Key()
		valueRaw := item.Leaf()

		mapItem := bpfMaps.LimitBandMapItem{
			Key:   bpf.GetHostByteOrder().Uint32(keyRaw),
			Value: bpf.GetHostByteOrder().Uint32(valueRaw),
		}
		result = append(result, mapItem)

	}
	fmt.Printf("peers in map %s: %d\n", bpfMaps.LimitBandMap, result)
	return result, nil
}
