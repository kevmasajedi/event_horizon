package workers

import (
	"encoding/json"
	"errors"
	"event_horizon/system/db"
	"event_horizon/system/hub"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
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

func AssertGEQ(hub *hub.Hub, trigger string, emission string, negative_emission string, key1 string, key2 string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			value1, err1 := RetrieveNestedValue(hub.Context(), key1)
			if err1 != nil {
				fmt.Println(err1.Error())
				hub.RedLink() <- "KEY1_NOT_FOUND"
				return
			}
			value2, err2 := RetrieveNestedValue(hub.Context(), key2)
			if err2 != nil {
				hub.RedLink() <- "KEY2_NOT_FOUND"
				return
			}
			var num1, num2 int
			var err error

			// Function to convert string with commas to int
			convertStringToInt := func(s string) (int, error) {
				s = strings.ReplaceAll(s, ",", "") // Remove commas
				return strconv.Atoi(s)
			}
			switch v1 := value1.(type) {
			case int:
				num1 = v1
			case string:
				num1, err = convertStringToInt(v1)
				if err != nil {
					hub.RedLink() <- "CONV_ERR_KEY_1"
					return
				}
			default:
				hub.RedLink() <- "CONV_ERR_KEY_1"
				return
			}
			switch v2 := value2.(type) {
			case int:
				num2 = v2
			case string:
				num2, err = convertStringToInt(v2)
				if err != nil {
					hub.RedLink() <- "CONV_ERR_KEY_2"
					return
				}
			default:
				hub.RedLink() <- "CONV_ERR_KEY_2"
				return
			}
			if num1 >= num2 {
				hub.LogLink() <- trigger + "->" + emission
				hub.UpLink() <- emission
			} else {
				hub.LogLink() <- trigger + "->" + negative_emission
				hub.UpLink() <- negative_emission
			}
		}
	}

}
func AssertEQ(hub *hub.Hub, trigger string, emission string, negative_emission string, key1 string, key2 string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			value1, err1 := RetrieveNestedValue(hub.Context(), key1)
			if err1 != nil {
				fmt.Println(err1.Error())
				hub.RedLink() <- "KEY1_NOT_FOUND"
				return
			}
			value2, err2 := RetrieveNestedValue(hub.Context(), key2)
			if err2 != nil {
				hub.RedLink() <- "KEY2_NOT_FOUND"
				return
			}

			if value1 == value2 {
				hub.LogLink() <- trigger + "->" + emission
				hub.UpLink() <- emission
			} else {
				hub.LogLink() <- trigger + "->" + negative_emission
				hub.UpLink() <- negative_emission
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
				if ch == ',' {
					continue
				}
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
func LoadContextFromCollectionIntoKey(hub *hub.Hub, trigger string, emission string, collection_name string, dest_key string, id_mask_key string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			result := db.FindOneFromCollection(collection_name, map[string]interface{}{"id": hub.Context()[id_mask_key]})
			if result != nil {
				hub.Context()[dest_key] = result

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
func SetContextKeyAsDirectValue(hub *hub.Hub, trigger string, emission string, key string, value interface{}) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			hub.Context()[key] = value

			hub.LogLink() <- trigger + "->" + emission
			hub.UpLink() <- emission
		}
	}
}

func SetContextKeyAsFormattedSumOfArray(hub *hub.Hub, trigger string, emission string, input_array_key string, output_key string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			inputArray, exists := hub.Context()[input_array_key].([]string)
			if !exists {
				hub.RedLink() <- "INPUT_ARRAY_NOT_FOUND"
				return
			}

			sum := 0
			for _, elem := range inputArray {
				// Remove commas and convert to integer
				cleanedElem := strings.ReplaceAll(elem, ",", "")
				if num, err := strconv.Atoi(cleanedElem); err == nil {
					sum += num
				} else {
					hub.RedLink() <- "INVALID_NUMBER_IN_ARRAY"
					return
				}
			}

			// Format the sum with commas
			formattedSum := fmt.Sprintf("%d", sum)
			formattedSumWithCommas := addCommas(formattedSum)

			// Set the formatted sum in the context
			hub.Context()[output_key] = formattedSumWithCommas
			hub.LogLink() <- trigger + "->" + emission
			hub.UpLink() <- emission
		}
	}
}
func SetContextKeyAsFormattedMultipleOfKey(hub *hub.Hub, trigger string, emission string, key string, coefficient float64, output_key string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			valStr, exists := hub.Context()[key].(string)
			if !exists {
				hub.RedLink() <- "VALUE_NOT_FOUND"
				return
			}

			cleanedVal := strings.ReplaceAll(valStr, ",", "")
			val, err := strconv.ParseFloat(cleanedVal, 64)
			if err != nil {
				hub.RedLink() <- "INVALID_NUMBER"
				return
			}

			result := val * coefficient

			formattedResult := fmt.Sprintf("%.0f", result)
			formattedResultWithCommas := addCommas(formattedResult)

			hub.Context()[output_key] = formattedResultWithCommas
			hub.LogLink() <- trigger + "->" + emission
			hub.UpLink() <- emission
		}
	}
}
func SetContextKeyAsFormattedSumOfKeys(hub *hub.Hub, trigger string, emission string, key1 string, key2 string, output_key string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			// Retrieve the values from the context and convert to float
			valStr1, exists1 := hub.Context()[key1].(string)
			valStr2, exists2 := hub.Context()[key2].(string)

			if !exists1 || !exists2 {
				hub.RedLink() <- "VALUE_NOT_FOUND"
				return
			}

			// Remove commas from the values
			cleanedVal1 := strings.ReplaceAll(valStr1, ",", "")
			cleanedVal2 := strings.ReplaceAll(valStr2, ",", "")

			val1, err1 := strconv.ParseFloat(cleanedVal1, 64)
			val2, err2 := strconv.ParseFloat(cleanedVal2, 64)

			if err1 != nil || err2 != nil {
				hub.RedLink() <- "INVALID_NUMBER"
				return
			}

			// Add the values
			result := val1 + val2

			// Format the result with commas
			formattedResult := fmt.Sprintf("%.0f", result)
			formattedResultWithCommas := addCommas(formattedResult)

			// Set the formatted result in the context
			hub.Context()[output_key] = formattedResultWithCommas
			hub.LogLink() <- trigger + "->" + emission
			hub.UpLink() <- emission
		}
	}
}
func SetContextKeyAsFormattedDiffOfKeys(hub *hub.Hub, trigger string, emission string, key1 string, key2 string, output_key string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			// Retrieve the values from the context using the helper function
			val1Raw, err1 := RetrieveNestedValue(hub.Context(), key1)
			val2Raw, err2 := RetrieveNestedValue(hub.Context(), key2)

			if err1 != nil || err2 != nil {
				hub.RedLink() <- "VALUE_NOT_FOUND"
				return
			}

			// Convert the retrieved values to strings
			valStr1, ok1 := val1Raw.(string)
			valStr2, ok2 := val2Raw.(string)

			if !ok1 || !ok2 {
				hub.RedLink() <- "INVALID_VALUE_TYPE"
				return
			}

			// Remove commas from the values
			cleanedVal1 := strings.ReplaceAll(valStr1, ",", "")
			cleanedVal2 := strings.ReplaceAll(valStr2, ",", "")

			val1, err1 := strconv.ParseInt(cleanedVal1, 10, 64)
			val2, err2 := strconv.ParseInt(cleanedVal2, 10, 64)

			if err1 != nil || err2 != nil {
				hub.RedLink() <- "INVALID_NUMBER"
				return
			}

			// Calculate the difference
			result := val1 - val2

			// Format the result with commas
			formattedResult := fmt.Sprintf("%d", result)
			formattedResultWithCommas := addCommas(formattedResult)

			// Set the formatted result in the context
			hub.Context()[output_key] = formattedResultWithCommas
			hub.LogLink() <- trigger + "->" + emission
			hub.UpLink() <- emission
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
func AppendObjectToArray(hub *hub.Hub, trigger string, emission string, value_key string, array_key string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			if _, exists := hub.Context()[array_key]; !exists {
				hub.Context()[array_key] = []interface{}{hub.Context()[value_key]}
			} else {
				arr := hub.Context()[array_key].(primitive.A)
				hub.Context()[array_key] = append(arr, hub.Context()[value_key])
			}
			hub.LogLink() <- trigger + "->" + emission
			hub.UpLink() <- emission
		}
	}
}

func AppendDistinctValueToArray(hub *hub.Hub, trigger string, emission string, negative_emission string, value_key string, array_key string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			// Check if the array exists in the context
			if _, exists := hub.Context()[array_key]; !exists {
				// Initialize the array if it doesn't exist and add the value
				hub.Context()[array_key] = []string{hub.Context()[value_key].(string)}

				hub.LogLink() <- trigger + "->" + emission
				hub.UpLink() <- emission
			} else {
				// Convert existing array to a slice of strings
				arr, err := convertPrimitiveAToStringSlice(hub.Context()[array_key].(primitive.A))
				if err != nil {
					fmt.Println(err.Error())
					hub.RedLink() <- "APPEND_VAL_ERR"
				}

				// Check if the value is already in the array
				value := hub.Context()[value_key].(string)
				contains := false
				for _, v := range arr {
					if v == value {
						contains = true
						break
					}
				}

				// Only append the value if it isn't already in the array
				if !contains {
					hub.Context()[array_key] = append(arr, value)

					hub.LogLink() <- trigger + "->" + emission
					hub.UpLink() <- emission
				} else {
					hub.LogLink() <- trigger + "->" + negative_emission
					hub.UpLink() <- negative_emission
				}
			}
		}
	}
}
func CopyContextKey(hub *hub.Hub, trigger string, emission string, key_1 string, key_2 string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			hub.Context()[key_2] = hub.Context()[key_1]
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
func UpsertContextAsItemIntoCollection(hub *hub.Hub, trigger string, emission string, except_keys []string, collection_name string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			newItem := make(map[string]interface{}, len(hub.Context()))
			for key, value := range hub.Context() {
				newItem[key] = value
			}
			for _, k := range except_keys {
				delete(newItem, k)
			}
			if db.UpsertItemInCollection(collection_name, newItem, "id") {
				hub.LogLink() <- trigger + "->" + emission
				hub.UpLink() <- emission
			}
		}
	}
}
func DeleteItemFromCollection(hub *hub.Hub, trigger string, emission string, context_keys []string, collection_name string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			filter := make(map[string]interface{})
			for _, k := range context_keys {
				filter[k] = hub.Context()[k]
			}
			if db.DeleteOneFromCollection(collection_name, filter) {
				hub.LogLink() <- trigger + "->" + emission
				hub.UpLink() <- emission
			}
		}
	}
}
func MapArrayElementsToCollectionValue(hub *hub.Hub, trigger string, emission string, input_array_key string, collection_name string, output_array_key string, result_key string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			// Check if input array exists in hub context as primitive.A
			inputArray, exists := hub.Context()[input_array_key].(primitive.A)
			if !exists {
				hub.RedLink() <- "INPUT_ARRAY_NOT_FOUND"
				return
			}

			outputArray := []string{}

			// Loop through each element in the input array and perform a lookup in the collection
			for _, elem := range inputArray {
				// Make sure elem is string or properly cast it
				if elemStr, ok := elem.(string); ok {
					result := db.FindOneFromCollection(collection_name, map[string]interface{}{"id": elemStr})
					if result != nil {
						// Extract the desired value from the result using result_key
						if value, found := result[result_key].(string); found {
							outputArray = append(outputArray, value)
						}
					}
				}
			}

			// Store the results in the output array key
			hub.Context()[output_array_key] = outputArray

			// Log and emit the trigger and emission
			hub.LogLink() <- trigger + "->" + emission
			hub.UpLink() <- emission
		}
	}
}

func MapArrayElementsToCollectionResults(hub *hub.Hub, trigger string, emission string, input_array_key string, collection_name string, output_array_key string) {
	for msg := range hub.DownLink() {
		if msg == trigger {
			// Check if input array exists in hub context as primitive.A
			inputArray, exists := hub.Context()[input_array_key].(primitive.A)
			if !exists {
				hub.RedLink() <- "INPUT_ARRAY_NOT_FOUND"
				return
			}

			outputArray := []interface{}{}

			// Loop through each element in the input array and perform a lookup in the collection
			for _, elem := range inputArray {
				// Make sure elem is string or properly cast it
				if elemStr, ok := elem.(string); ok {
					result := db.FindOneFromCollection(collection_name, map[string]interface{}{"id": elemStr})
					if result != nil {
						outputArray = append(outputArray, result)
					}
				}
			}

			// Store the results in the output array key
			hub.Context()[output_array_key] = outputArray

			// Log and emit the trigger and emission
			hub.LogLink() <- trigger + "->" + emission
			hub.UpLink() <- emission
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

// Helper function to add commas to the number
func addCommas(s string) string {
	n := len(s)
	if n <= 3 {
		return s
	}
	rem := n % 3
	if rem == 0 {
		rem = 3
	}
	return s[:rem] + "," + strings.Join(splitEvery(s[rem:], 3), ",")
}

// Helper function to split the string every n characters
func splitEvery(s string, n int) []string {
	result := []string{}
	for len(s) > 0 {
		if len(s) <= n {
			result = append(result, s)
			break
		}
		result = append(result, s[:n])
		s = s[n:]
	}
	return result
}

func RetrieveNestedValue(context map[string]interface{}, key string) (interface{}, error) {
	keys := strings.Split(key, ".")  // Split the key into parts
	var result interface{} = context // Start with the top-level context
	for _, k := range keys {
		switch res := result.(type) {
		case map[string]interface{}:
			if val, exists := res[k]; exists {
				result = val
			} else {
				return nil, errors.New("key not found")
			}
		default:
			return nil, errors.New("invalid type encountered")
		}
	}

	return result, nil
}
