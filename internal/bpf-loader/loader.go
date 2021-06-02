package bpf_loader

import (
	"encoding/binary"
	"fmt"
	bpf "github.com/iovisor/gobpf/bcc"
	"net"
)

/*
#cgo CFLAGS: -I/usr/include/bcc/compat
#cgo LDFLAGS: -lbcc
#include <bcc/bcc_common.h>
#include <bcc/libbpf.h>
*/
import "C"

func LoadModule(ip net.IP) *bpf.Module {
	ipToInt32 := fmt.Sprintf("%d", binary.LittleEndian.Uint32(ip))

	return bpf.NewModule(CSourceCode, []string{
		"-w",
		"-DLOCAL_ADDR=" + ipToInt32,
	},
	)
}
