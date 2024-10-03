package workers

import (
	"encoding/json"
	"event_horizon/system/db"
	"event_horizon/system/hub"
	"fmt"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func PhoneNumberValidator(hub *hub.Hub, trigger string, emission string, key string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			val := hub.Context()[key].(string)
			if len(val) == 11 && val[:2] == "09" {
				hub.LogLink() <- trigger + "->" + emission
				hub.UpLink() <- emission
			} else {
				hub.RedLink() <- "INVALID_PHONE_NUMBER"
			}
		}
	}
}

func NumericSanitizer(hub *hub.Hub, trigger string, emission string, key string) {
	persianArabicToEnglish := map[rune]rune{
		'۰': '0', '١': '1', '٢': '2', '٣': '3', '٤': '4', '٥': '5', '٦': '6', '٧': '7', '٨': '8', '٩': '9',
		'۱': '1', '۲': '2', '۳': '3', '۴': '4', '۵': '5', '۶': '6', '۷': '7', '۸': '8', '۹': '9',
	}

	for msg := range hub.DownLink() {
		if msg == trigger {
			val := hub.Context()[key].(string)
			normalized := []rune{}
			isNumeric := true

			for _, ch := range val {
				if englishChar, exists := persianArabicToEnglish[ch]; exists {
					normalized = append(normalized, englishChar)
				} else if ch >= '0' && ch <= '9' {
					normalized = append(normalized, ch)
				} else {
					isNumeric = false
					break
				}
			}

			if isNumeric {
				hub.Context()[key] = string(normalized) // Store the normalized value back
				hub.LogLink() <- trigger + "->" + emission
				hub.UpLink() <- emission
			} else {
				hub.RedLink() <- "NON_NUMERIC_CHARACTER_FOUND"
			}
		}
	}
}

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
func LoadOptionalContextFromCollection(hub *hub.Hub, trigger string, emission string, negative_emission string, collection_name string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			result := db.FindOneFromCollection(collection_name, hub.Context())
			if result != nil {
				for key, value := range result {
					hub.Context()[key] = value
				}
				hub.LogLink() <- trigger + "->" + emission
				hub.UpLink() <- emission
			} else {
				hub.LogLink() <- trigger + "->" + negative_emission
				hub.UpLink() <- negative_emission
			}
		}
	}
}
func LoadContextFromCollectionByKeys(hub *hub.Hub, trigger string, emission string, collection_name string, keys []string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			ctx := make(map[string]interface{})
			for _, k := range keys {
				ctx[k] = hub.Context()[k]
			}
			result := db.FindOneFromCollection(collection_name, ctx)
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
func LoadOptionalContextFromCollectionWithKeys(hub *hub.Hub, trigger string, emission string, negative_emission string, collection_name string, keys []string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			ctx := make(map[string]interface{})
			for _, k := range keys {
				ctx[k] = hub.Context()[k]
			}
			result := db.FindOneFromCollection(collection_name, ctx)
			if result != nil {
				for key, value := range result {
					hub.Context()[key] = value
				}
				hub.LogLink() <- trigger + "->" + emission
				hub.UpLink() <- emission
			} else {
				hub.LogLink() <- trigger + "->" + negative_emission
				hub.UpLink() <- negative_emission
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
func DumpContextKeysAsJSON(hub *hub.Hub, trigger string, keys []string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			ctx := make(map[string]interface{})
			for _, k := range keys {
				ctx[k] = hub.Context()[k]
			}
			jsonData, err := json.Marshal(ctx)
			if err != nil {
				hub.RedLink() <- "JSON_MARSHAL_ERR"
			} else {
				hub.RedLink() <- string(jsonData)
			}
		}
	}
}
func IsKeySupplied(hub *hub.Hub, trigger string, emission string, negative_emission string, key string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			_, exists := hub.Context()[key]
			if exists {
				hub.LogLink() <- trigger + "->" + emission
				hub.UpLink() <- emission
			} else {
				hub.LogLink() <- trigger + "->" + negative_emission
				hub.UpLink() <- negative_emission
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
func AppendValueToArray(hub *hub.Hub, trigger string, emission string, value_key string, array_key string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			if _, exists := hub.Context()[array_key]; !exists {
				hub.Context()[array_key] = []string{hub.Context()[value_key].(string)}
			} else {
				arr, err := convertPrimitiveAToStringSlice(hub.Context()[array_key].(primitive.A))
				if err != nil {
					fmt.Println(err.Error())
					hub.RedLink() <- "APPEND_VAL_ERR"
				}
				hub.Context()[array_key] = append(arr, hub.Context()[value_key].(string))
			}
			hub.LogLink() <- trigger + "->" + emission
			hub.UpLink() <- emission
		}
	}
}
func RemoveValueFromArray(hub *hub.Hub, trigger string, emission string, value_key string, array_key string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			if arr, exists := hub.Context()[array_key]; exists {
				arrSlice, err := convertPrimitiveAToStringSlice(arr.(primitive.A))
				if err != nil {
					hub.RedLink() <- "REMOVE_VAL_ERR"
				} else {
					valueToRemove := hub.Context()[value_key].(string)
					newArr := []string{}
					removed := false

					for _, val := range arrSlice {
						if val == valueToRemove && !removed {
							removed = true
							continue
						}
						newArr = append(newArr, val)
					}

					hub.Context()[array_key] = newArr
				}
			}
			hub.LogLink() <- trigger + "->" + emission
			hub.UpLink() <- emission
		}
	}
}
func UpsertKeysAsItemIntoCollection(hub *hub.Hub, trigger string, emission string, context_keys []string, collection_name string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			newItem := make(map[string]interface{})
			for _, k := range context_keys {
				newItem[k] = hub.Context()[k]
			}
			if db.UpsertItemInCollection(collection_name, newItem, "id") {
				hub.LogLink() <- trigger + "->" + emission
				hub.UpLink() <- emission
			}
		}
	}
}
func GenerateUniqueId(hub *hub.Hub, trigger string, emission string, as_key string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			b := make([]byte, 8)
			for i := range b {
				b[i] = charset[rand.Intn(len(charset))]
			}
			hub.Context()[as_key] = string(b)
			hub.LogLink() <- trigger + "->" + emission
			hub.UpLink() <- emission
		}
	}
}
func TimeStampGenerator(hub *hub.Hub, trigger string, emission string, as_key string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			// Get the current timestamp in a suitable format (e.g., RFC3339)
			timestamp := time.Now().Format(time.RFC3339)

			// Store the timestamp in the hub's context
			hub.Context()[as_key] = timestamp

			// Log and emit the trigger and emission
			hub.LogLink() <- trigger + "->" + emission
			hub.UpLink() <- emission
		}
	}
}
func TwoFactorCodeGenerator(hub *hub.Hub, trigger string, emission string, key string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			code := fmt.Sprintf("%06d", rand.Intn(1000000))
			hub.Context()[key] = code
			hub.LogLink() <- trigger + "->" + emission
			hub.UpLink() <- emission
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

func convertPrimitiveAToStringSlice(arr primitive.A) ([]string, error) {
	// Create a slice to hold the string values
	strSlice := make([]string, len(arr))

	// Loop through the primitive.A array
	for i, v := range arr {
		// Try to assert the type of each element as string
		if str, ok := v.(string); ok {
			strSlice[i] = str
		} else {
			return nil, fmt.Errorf("element at index %d is not a string: %v", i, v)
		}
	}

	return strSlice, nil
}
