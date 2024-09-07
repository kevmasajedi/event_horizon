package main

import (
	"event_horizon/system/autoinvoker"
	"event_horizon/system/domains"
	"event_horizon/workers"
)

func main() {
	var context map[string]string
	autoinvoker.InitializeContext("domain_name", &context, []string{"phone_number"}, true)

	domains.InitializeDomain(context, domain_workers_bootstrapper)
	domains.Run("impulse_in", "success")
}

func domain_workers_bootstrapper(backlink chan string, loglink chan string, redlink chan string, domain_context map[string]string) {
	go workers.NilWorker(domains.NewDedicatedLink(), backlink, loglink, redlink, "impulse_in", "success", "error_validating")
}
