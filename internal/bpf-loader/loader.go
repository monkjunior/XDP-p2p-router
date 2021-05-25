package bpf_loader

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	bpf "github.com/iovisor/gobpf/bcc"
)

/*
#cgo CFLAGS: -I/usr/include/bcc/compat
#cgo LDFLAGS: -lbcc
#include <bcc/bcc_common.h>
#include <bcc/libbpf.h>
*/
import "C"

func LoadModule(device string) *bpf.Module {
	ip, err := getLocalIP(device)
	if err != nil {
		fmt.Printf("failed to get local ip: %v\n", err)
	}
	fmt.Println("your local ip address", ip.String())

	ipToInt32 := fmt.Sprintf("%d", binary.LittleEndian.Uint32(ip))
	fmt.Println("your local ip in format int32: ", ipToInt32)

	return bpf.NewModule(CSourceCode, []string{
		"-w",
		"-DLOCAL_ADDR=" + ipToInt32,
	},
	)
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
	addresses, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addresses {
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
