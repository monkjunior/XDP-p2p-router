package common

import (
	"fmt"
	"io/ioutil"
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
	fmt.Println("my public ip:", IP)
	return IP, nil
}
