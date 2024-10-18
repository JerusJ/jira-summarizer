package util

import (
	"fmt"
	"os"
)

func GetEnvOrDie(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic(fmt.Sprintf("Could not get environment variable: '%s'", k))
	}
	return v
}
