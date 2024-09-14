package system

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var logs string = ""
var should_log bool = false

func Logger(log_link chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for edge := range log_link {
		logs = fmt.Sprintf(logs+"%s\n", edge)
		if should_log {
			fmt.Println(logs)
		}
	}
}
func TurnOnLogger() {
	should_log = true
}
func Panicker(panic_link chan string) {
	for msg := range panic_link {
		fmt.Printf("%s", msg)
		os.Exit(1)
	}
}
func Broadcaster(from chan string, to []chan string) {
	for msg := range from {
		for _, ch := range to {
			go func() {
				ch <- msg
			}()
		}
	}
}
func Repeater(downlink chan string, uplink chan string, initiator string, terminator string) {
	to := time.After(8 * time.Second)
	uplink <- initiator
	for {
		select {
		case resp := <-downlink:
			if resp == terminator {
				return
			}
			uplink <- resp
			to = time.After(10 * time.Second)
		case <-to:
			fmt.Println("TimeOut!")
			os.Exit(0)
			return
		}
	}
}
