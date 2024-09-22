package main

import (
	// "event_horizon/system"
	"event_horizon/system/autoinvoker"
	"event_horizon/system/domains"
	"event_horizon/system/hub"
	"event_horizon/workers"
)

func main() {
	var context map[string]interface{}

	autoinvoker.AutoInitialize("product_list_view", &context, []string{"id"}, "local")
	// system.TurnOnLogger()
	domains.InitializeDomain(context, slots, domain_workers_bootstrapper)
	domains.Run("impulse_in", "chain_ended")
}

func domain_workers_bootstrapper() {
	go workers.NilWorker(hub.NewHub(), "impulse_in", "say_hello")
	go workers.SayHello(hub.NewHub(), "say_hello", "chain_ended", "name")
}
