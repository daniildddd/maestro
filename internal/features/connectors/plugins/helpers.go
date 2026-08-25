package plugins

import (
	"strings"
)

func firstEntry(list string) string {
	first := list

	if index := strings.IndexByte(first, ','); index >= 0 {
		first = first[:index]
	}

	return first
}

func splitHostPort(hostport string) (host, port string) {
	host, port = hostport, ""

	if index := strings.LastIndexByte(hostport, ':'); index >= 0 {
		host, port = hostport[:index], hostport[index+1:]
	}

	return host, port
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}
