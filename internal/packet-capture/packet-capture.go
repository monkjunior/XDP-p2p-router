package packet_capture

import (
	"fmt"
	bpf "github.com/iovisor/gobpf/bcc"
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
		Table: bpf.NewTable(module.TableId("counters"), module),
	}, nil
}

func Close(device string, module *bpf.Module){
	if err := module.RemoveXDP(device); err != nil {
		fmt.Fprintf(os.Stderr, "failed to remove XDP from %s: %v\n", device, err)
	}
	fmt.Println("close packet capture module successfully")
}

func (p *PacketCapture) PrintCounterMap(){
	countersTable := p.Table
	fmt.Println("counter table", countersTable, countersTable.Config())
	fmt.Printf("\n{Keys}: {Values}\n")
	for item := countersTable.Iter(); item.Next(); {
		keyRaw := item.Key()
		valueRaw := item.Leaf()
		key := PacketInfo{
			SourceAddr: keyRaw[0:4],
			DestAddr:   keyRaw[4:8],
			Family:     keyRaw[8:12],
		}
		value := PacketCounter{
			RxPackets: valueRaw[0:4],
			RxBytes:   valueRaw[4:8],
		}
		fmt.Printf("%+v: %+v\n", key, value)
	}
}


