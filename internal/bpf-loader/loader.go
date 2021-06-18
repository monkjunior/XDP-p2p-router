package bpf_loader

import (
	"encoding/binary"
	"fmt"
	bpf "github.com/iovisor/gobpf/bcc"
	"github.com/spf13/viper"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/common"
	"net"
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

func LoadModule(ip net.IP) *bpf.Module {
	blockedIP, _ := common.ConvertIPToUint32(viper.GetString("blocked_address"))
	return bpf.NewModule(CSourceCode, []string{
		"-w",
		"-DLOCAL_ADDR=" + fmt.Sprintf("%d", binary.LittleEndian.Uint32(ip)),
		"-DBLOCK_ADDR=" + fmt.Sprintf("%d", blockedIP),
	},
	)
}
