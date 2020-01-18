package openvvar

import (
	"fmt"
	"os"
	"strings"
)

type NotFound struct {}

func (nf *NotFound) Error() string {
	return fmt.Sprint("env var not found")
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