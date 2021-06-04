package packet_capture

import (
	"fmt"
	"os"

	bpf "github.com/iovisor/gobpf/bcc"
	bpfMaps "github.com/vu-ngoc-son/XDP-p2p-router/internal/bpf-maps"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/common"
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
	fn, err := module.Load("packet_counter", C.BPF_PROG_TYPE_XDP, 1, 65536)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load xdp program packetCounter: %v\n", err)
		return nil, err
	}

	err = module.AttachXDP(device, fn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to attach xdp prog: %v\n", err)
		return nil, err
	}

	return &PacketCapture{
		Table: bpf.NewTable(module.TableId(bpfMaps.PacketCaptureMap), module),
	}, nil
}

func Close(device string, module *bpf.Module) {
	if err := module.RemoveXDP(device); err != nil {
		fmt.Fprintf(os.Stderr, "failed to remove XDP from %s: %v\n", device, err)
	}
	fmt.Println("close packet capture module successfully")
}

func (p *PacketCapture) ExportMap() (result []bpfMaps.PktCounterMapItem, err error) {
	countersTable := p.Table

	result = make([]bpfMaps.PktCounterMapItem, 0)
	for item := countersTable.Iter(); item.Next(); {
		if item.Err() != nil {
			continue
		}

		keyRaw := item.Key()
		valueRaw := item.Leaf()

		rxPackets, err := common.ConvertUint8ToUInt64(valueRaw[0:8])
		if err != nil {
			return nil, err
		}

		rxBytes, err := common.ConvertUint8ToUInt64(valueRaw[8:16])
		if err != nil {
			return nil, err
		}

		mapItem := bpfMaps.PktCounterMapItem{
			Key: keyRaw,
			Value: bpfMaps.PktCounterValue{
				RxPackets: rxPackets,
				RxBytes:   rxBytes,
			},
		}
		result = append(result, mapItem)

	}
	return result, nil
}
