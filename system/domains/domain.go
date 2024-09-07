package domains

import (
	"event_horizon/system"
	"sync"
	"time"
)

var DomainDownlink chan string
var DomainUplink chan string
var DomainLogChannel chan string
var DomainRedChannel chan string

var DomainDedicatedLinks []chan string

var DomainContext map[string]string

var DomainBootFunction func(chan string, chan string, chan string, map[string]string)

var DomainWg sync.WaitGroup

func InitializeDomain(context map[string]string, bootWith func(chan string, chan string, chan string, map[string]string)) {

	DomainDownlink = make(chan string)
	DomainUplink = make(chan string)
	DomainLogChannel = make(chan string)
	DomainRedChannel = make(chan string)
	SetDomainContext(context)
	SetDomainBootFunction(bootWith)
	Boot()
}
func GetDownlink() chan string {
	return DomainDownlink
}
func GetUplink() chan string {
	return DomainUplink
}
func GetLogChannel() chan string {
	return DomainLogChannel
}
func GetRedChannel() chan string {
	return DomainRedChannel
}
func NewDedicatedLink() chan string {
	newLink := make(chan string)
	DomainDedicatedLinks = append(DomainDedicatedLinks, newLink)
	return newLink
}
func GetDomainDedicatedLinks() []chan string {
	return DomainDedicatedLinks
}
func GetDomainContext() map[string]string {
	return DomainContext
}
func SetDomainContext(dc map[string]string) {
	DomainContext = dc
}
func SetDomainBootFunction(f func(chan string, chan string, chan string, map[string]string)) {
	DomainBootFunction = f
}
func Boot() {
	DomainBootFunction(DomainDownlink, DomainLogChannel, DomainRedChannel, DomainContext)
	DomainWg.Add(1)
	go system.Logger(DomainLogChannel, &DomainWg)
	go system.Panicker(DomainRedChannel)
	go system.Broadcaster(DomainUplink, DomainDedicatedLinks)
}
func Run(initiator string, terminator string) {
	system.Repeater(DomainDownlink, DomainUplink, initiator, terminator)
	time.Sleep(1 * time.Millisecond)
	CloseLogChannel()
	DomainWg.Wait()
	CloseDedicatedLinks()
	CloseAllChannels()
}
func CloseLogChannel() {
	close(DomainLogChannel)
}
func CloseAllChannels() {
	close(DomainDownlink)
	close(DomainUplink)
}
func CloseDedicatedLinks() {
	for _, link := range DomainDedicatedLinks {
		close(link)
	}
	DomainDedicatedLinks = nil
}
