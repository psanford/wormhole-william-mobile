package plog

import (
	"fmt"
	"log"
	"time"
)

var (
	pendingMsg = make(chan string, 100)
)

func Printf(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	log.Print(str)
	msg := fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC3339), str)

	select {
	case pendingMsg <- msg:
	default:
	}
}

func MsgChan() chan string {
	return pendingMsg
}
