package common

import (
	"errors"
	"io/ioutil"
	"net"
	"net/http"
)

func GetMyPublicIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			return
		}
	}()

	IP := string(bodyBytes)
	return IP, nil
}

func GetMyPrivateIP(device string) (net.IP, error) {
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
