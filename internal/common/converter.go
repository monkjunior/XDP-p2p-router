package common

import (
	"fmt"
	"net"

	"github.com/iovisor/gobpf/bcc"
)

//ConvertUint8ToIP quickly convert raw data to string format of IP without doing any validation
func ConvertUint8ToIP(ipRaw []uint8) (string, error) {
	if len(ipRaw) != 4 {
		return "", ErrFailedToConvertIP
	}
	IP := fmt.Sprintf("%d.%d.%d.%d", ipRaw[0], ipRaw[1], ipRaw[2], ipRaw[3])
	if IP == "" {
		return "", ErrEmptyResultConvertIP
	}
	return IP, nil
}

func ConvertUint8ToUInt32(rawData []uint8) (uint32, error) {
	if len(rawData) != 4 {
		return 0, ErrFailedToConvertUint32
	}
	result := bcc.GetHostByteOrder().Uint32(rawData)
	return result, nil
}

func ConvertUint8ToUInt64(rawData []uint8) (uint64, error) {
	if len(rawData) != 8 {
		return 0, ErrFailedToConvertUint64
	}
	result := bcc.GetHostByteOrder().Uint64(rawData)
	return result, nil
}

func ConvertIPToUint32(ipStr string) (uint32, error) {
	IP := net.ParseIP(ipStr)
	if IP == nil {
		return 0, ErrWrongIPFormat
	}
	IP = IP.To4()
	return bcc.GetHostByteOrder().Uint32(IP), nil
}
