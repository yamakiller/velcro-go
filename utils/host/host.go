package host

import (
	"errors"
	"strconv"
	"strings"
)

var (
	ErrorHostAddressFormat = errors.New("Host address format error")
)

func Parse(addr string) (string, int, error) {
	hosts := strings.Split(addr, ":")
	if len(hosts) != 2 {
		return "", 0, ErrorHostAddressFormat
	}

	ip := hosts[0]
	port, err := strconv.ParseInt(hosts[1], 10, 32)
	if err != nil {
		return "", 0, err
	}

	return ip, int(port), nil
}
