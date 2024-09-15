package hub

import "event_horizon/system/domains"

type HubInterface interface {
	DownLink() chan string
	UpLink() chan string
	LogLink() chan string
	RedLink() chan string
	Context() map[string]string
}

type Hub struct {
	downLink chan string
	upLink   chan string
	logLink  chan string
	redLink  chan string
	context  map[string]string
}

func NewHub() *Hub {
	return &Hub{
		downLink: domains.NewDedicatedLink(),
		upLink:   domains.GetDownlink(), // hub's up link is domain's down link
		logLink:  domains.GetLogChannel(),
		redLink:  domains.GetRedChannel(),
		context:  domains.GetDomainContext(),
	}
}
func (h *Hub) DownLink() chan string {
	return h.downLink
}

func (h *Hub) UpLink() chan string {
	return h.upLink
}

func (h *Hub) LogLink() chan string {
	return h.logLink
}

func (h *Hub) RedLink() chan string {
	return h.redLink
}

func (h *Hub) Context() map[string]string {
	return h.context
}
