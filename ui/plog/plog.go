package plog

import (
	"fmt"
	"log"
)

func Printf(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	log.Print(str)
}
