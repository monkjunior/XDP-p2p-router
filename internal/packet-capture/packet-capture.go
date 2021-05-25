package packet_capture

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"

	bpf "github.com/iovisor/gobpf/bcc"
)

/*
#cgo CFLAGS: -I/usr/include/bcc/compat
#cgo LDFLAGS: -lbcc
#include <bcc/bcc_common.h>
#include <bcc/libbpf.h>
*/
import "C"

func Start(device string) {
	ip, err := getLocalIP(device)
	if err != nil {
		fmt.Printf("failed to get local ip: %v\n", err)
	}

	fmt.Println("your local ip address", ip.String())

	ipToInt32 := fmt.Sprintf("%d", binary.LittleEndian.Uint32(ip))
	fmt.Println("your local ip in format int32: ", ipToInt32)
	module := bpf.NewModule(source, []string{
		"-w",
		"-DLOCAL_ADDR=" + ipToInt32,
	},
	)
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

func getLocalIP(device string) (net.IP, error) {
	iface, err := net.InterfaceByName(device)
	if err != nil {
		return nil, err
	}
	if iface.Flags&net.FlagUp == 0 {
		return nil, errors.New("your device is down")
	}
	if iface.Flags&net.FlagLoopback != 0 {
		return nil, errors.New("could not use loop back device")
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip == nil || ip.IsLoopback() {
			continue
		}
		ip = ip.To4()
		if ip == nil {
			continue // not an ipv4 address
		}
		return ip, nil
	}
	return nil, errors.New("could not connect to network")
}
