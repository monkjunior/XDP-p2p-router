package packet_capture

import (
	"fmt"
	bpf "github.com/iovisor/gobpf/bcc"
	bpf_maps "github.com/vu-ngoc-son/XDP-p2p-router/internal/bpf-maps"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/common"
	"os"
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
		Table: bpf.NewTable(module.TableId(bpf_maps.PacketCaptureMap), module),
	}, nil
}

func Close(device string, module *bpf.Module) {
	if err := module.RemoveXDP(device); err != nil {
		fmt.Fprintf(os.Stderr, "failed to remove XDP from %s: %v\n", device, err)
	}
	fmt.Println("close packet capture module successfully")
}

func (p *PacketCapture) ExportMap() (result []bpf_maps.PktCounterMapItem, err error) {
	countersTable := p.Table

	result = make([]bpf_maps.PktCounterMapItem, 0)
	for item := countersTable.Iter(); item.Next(); {
		keyRaw := item.Key()
		valueRaw := item.Leaf()

		sourceAddr, err := common.ConvertUint8ToIP(keyRaw[0:4])
		if err != nil {
			return nil, err
		}

		destAddr, err := common.ConvertUint8ToIP(keyRaw[4:8])
		if err != nil {
			return nil, err
		}

		family, err := common.ConvertUint8ToUInt32(keyRaw[8:12])
		if err != nil {
			return nil, err
		}

		rxPackets, err := common.ConvertUint8ToUInt64(valueRaw[0:8])
		if err != nil {
			return nil, err
		}

		rxBytes, err := common.ConvertUint8ToUInt64(valueRaw[8:16])
		if err != nil {
			return nil, err
		}

		mapItem := bpf_maps.PktCounterMapItem{
			Key: bpf_maps.PktCounterKey{
				SourceAddr: sourceAddr,
				DestAddr:   destAddr,
				Family:     family,
			},
			Value: bpf_maps.PktCounterValue{
				RxPackets: rxPackets,
				RxBytes:   rxBytes,
			},
		}
		result = append(result, mapItem)

	}
	fmt.Println(result)
	return result, nil
}
