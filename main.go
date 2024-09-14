package main

import (
	"event_horizon/system/autoinvoker"
	"event_horizon/system/domains"
	"event_horizon/workers"
)

func main() {
	var context map[string]string
	autoinvoker.AutoInitialize("hello", &context, []string{"name"}, "local")
	// system.TurnOnLogger()
	domains.InitializeDomain(context, domain_workers_bootstrapper)
	domains.Run("impulse_in", "chain_ended")
}

func domain_workers_bootstrapper(backlink chan string, loglink chan string, redlink chan string, domain_context map[string]string) {
	go workers.NilWorker(domains.NewDedicatedLink(), backlink, loglink, redlink, "impulse_in", "say_hello")
	go workers.SayHello(domains.NewDedicatedLink(), backlink, loglink, redlink, "say_hello", "chain_ended", domain_context, "name")
}
