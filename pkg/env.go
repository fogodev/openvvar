package openvvar

import (
	"os"
	"strings"
)

type NotFound struct{}

func (nf *NotFound) Error() string {
	return "env var not found"
}

func Get(name string) (string, *NotFound) {
	if val := os.Getenv(name); val != "" {
		return val, nil
	}
	name = strings.Replace(strings.ToUpper(name), "-", "_", -1)
	if val := os.Getenv(name); val != "" {
		return val, nil
	}
	return "", &NotFound{}
}
