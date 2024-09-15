package workers

import (
	"encoding/json"
	"event_horizon/system/db"
	"event_horizon/system/hub"
	"fmt"
)

func LoadContextFromCollection(hub *hub.Hub, trigger string, emission string, collection_name string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			result := db.FindOneFromCollection(collection_name, hub.Context())
			if result != nil {
				for key, value := range result {
					hub.Context()[key] = value
				}
				hub.LogLink() <- trigger + "->" + emission
				hub.UpLink() <- emission
			}
		}
	}
}
func DumpContextAsJSON(hub *hub.Hub, trigger string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			jsonData, err := json.Marshal(hub.Context())
			if err != nil {
				hub.RedLink() <- "JSON_MARSHAL_ERR"
			} else {
				hub.RedLink() <- string(jsonData)
			}
		}
	}
}
func InitDb(hub *hub.Hub, trigger string, emission string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			if db.Connect() {
				hub.LogLink() <- trigger + "->" + emission
				hub.UpLink() <- emission
			} else {
				hub.RedLink() <- "DB_INIT_ERROR"
			}
		}
	}
}

func NilWorker(hub *hub.Hub, trigger string, emission string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			hub.LogLink() <- trigger + "->" + emission
			hub.UpLink() <- emission
		}
	}
}
func SayHello(hub *hub.Hub, trigger string, emission string, name_field string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			hub.RedLink() <- fmt.Sprintf("Hello %s", hub.Context()[name_field])
		}
	}
}
