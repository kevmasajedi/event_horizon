package workers

import (
	"encoding/json"
	"event_horizon/system/db"
	"fmt"
)

func SaveContextToCollection(downlink chan string, uplink chan string, loglink chan string, trigger string, emission string, context map[string]string, collection_name string) {
	for msg := range downlink {
		if msg == trigger {
			if db.CreateCollection(collection_name) {
				db.InsertOneIntoCollection(collection_name, context)
				loglink <- trigger + "->" + emission
				uplink <- emission
			}
		}
	}
}
func LoadContextFromCollection(downlink chan string, uplink chan string, loglink chan string, trigger string, emission string, collection_name string, context map[string]string) {
	for msg := range downlink {
		if msg == trigger {
			result := db.FindOneFromCollection(collection_name, context)
			if result != nil {
				for key, value := range result {
					context[key] = value
				}
				loglink <- trigger + "->" + emission
				uplink <- emission
			}
		}
	}
}
func DumpContextAsJSON(downlink chan string, uplink chan string, redlink chan string, trigger string, context map[string]string) {
	for msg := range downlink {
		if msg == trigger {
			jsonData, err := json.Marshal(context)
			if err != nil {
				redlink <- "JSON_MARSHAL_ERR"
			} else {
				redlink <- string(jsonData)
			}
		}
	}
}
func InitDb(downlink chan string, uplink chan string, loglink chan string, redlink chan string, trigger string, emission string) {
	for msg := range downlink {
		if msg == trigger {
			if db.Connect() {
				loglink <- trigger + "->" + emission
				uplink <- emission
			} else {
				redlink <- "DB_INIT_ERROR"
			}
		}
	}
}

func NilWorker(downlink chan string, uplink chan string, loglink chan string, redlink chan string, trigger string, emission string) {
	for msg := range downlink {
		if msg == trigger {
			loglink <- trigger + "->" + emission
			uplink <- emission
		}
	}
}
func SayHello(downlink chan string, uplink chan string, loglink chan string, redlink chan string, trigger string, emission string, context map[string]string, name_field string) {
	for msg := range downlink {
		if msg == trigger {
			redlink <- fmt.Sprintf("Hello %s", context[name_field])
		}
	}
}
