package common

import "fmt"

//ConvertUint8ToIP quickly convert raw data to string format of IP without doing any validation
func ConvertUint8ToIP(ipRaw []uint8) (string, error) {
	if len(ipRaw) != 4 {
		return "", ErrFailedToConvertIP
	}
	return fmt.Sprintf("%d.%d.%d.%d", ipRaw[0], ipRaw[1], ipRaw[2], ipRaw[3]), nil
}
