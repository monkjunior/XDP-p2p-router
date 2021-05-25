package packet_capture

import (
	"fmt"

	bpf "github.com/iovisor/gobpf/bcc"
	"os"
	"os/signal"
)

/*
#cgo CFLAGS: -I/usr/include/bcc/compat
#cgo LDFLAGS: -lbcc
#include <bcc/bcc_common.h>
#include <bcc/libbpf.h>
*/
import "C"

func Start(device string) {
	module := bpf.NewModule(source, []string{"-w"})
	defer module.Close()

	fn, err := module.Load("packet_counter", C.BPF_PROG_TYPE_XDP, 1, 65536)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load xdp program packetCounter: %v\n", err)
		os.Exit(1)
	}

	err = module.AttachXDP(device, fn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to attach xdp prog: %v\n", err)
		os.Exit(1)
	}

	defer func() {
		if err := module.RemoveXDP(device); err != nil {
			fmt.Fprintf(os.Stderr, "failed to remove XDP from %s: %v\n", device, err)
		}
	}()

	fmt.Println("packet capturing, hit CTRL+C to stop")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	countersTable := bpf.NewTable(module.TableId("counters"), module)

	<-sig

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

	fmt.Println("shutting down gracefully ...")
}
