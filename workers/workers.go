package workers

import (
	"bytes"
	"encoding/json"
	"event_horizon/system/db"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"
)

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
func InitDb(downlink chan string, uplink chan string, loglink chan string, trigger string, emission string) {
	for msg := range downlink {
		if msg == trigger {
			if db.Connect() {
				loglink <- trigger + "->" + emission
				uplink <- emission
			} else {
				loglink <- trigger + "->" + "ERROR: could not initialize database!"
			}
		}
	}
}
func Send2FACode(downlink chan string, uplink chan string, loglink chan string, trigger string, emission string, context map[string]string, phone_slot string, code_slot string) {
	for msg := range downlink {
		if msg == trigger {
			url := "https://api2.ippanel.com/api/v1/sms/pattern/normal/send"
			apiKey := "t1pFhDwsmG_INnefQr73yjD9kP_1mrLosHdU6xInjl4="
			headers := map[string]string{
				"accept":       "application/json",
				"apikey":       apiKey,
				"Content-Type": "application/json",
			}
			data := map[string]interface{}{
				"code":      "4d9ckmeqywpotsw",
				"sender":    "+983000505",
				"recipient": context[phone_slot],
				"variable": map[string]string{
					"otp": context[code_slot],
				},
			}
			jsonData, _ := json.Marshal(data)
			req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

			for key, value := range headers {
				req.Header.Set(key, value)
			}
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				loglink <- trigger + "->" + "ERROR: could not send code!"
				continue
			}
			if resp.StatusCode == http.StatusOK {
				loglink <- trigger + "->" + emission
				uplink <- emission
			}
			resp.Body.Close()
		}
	}
}
func Generate2FACode(downlink chan string, uplink chan string, loglink chan string, trigger string, emission string, context map[string]string, slot string) {
	for msg := range downlink {
		if msg == trigger {
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			code := rng.Intn(1000000)
			codeStr := strconv.Itoa(code)
			codeStr = fmt.Sprintf("%06s", codeStr)
			context[slot] = codeStr
			loglink <- trigger + "->" + emission
			uplink <- emission
		}
	}
}

func ValidatePublicKey(downlink chan string, uplink chan string, loglink chan string, trigger string, emission string, context map[string]string, slot string) {
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="
	for msg := range downlink {
		if msg == trigger {
			public_key := context[slot]
			if len(public_key) != 68 {
				loglink <- trigger + "->" + "ERROR: could not validate public key!"
				return
			}
			for _, char := range public_key {
				if !strings.ContainsRune(base64Chars, char) {
					loglink <- trigger + "->" + "ERROR: could not validate public key!"
					return
				}
			}
			loglink <- trigger + "->" + emission
			uplink <- emission
		}
	}
}

func SanitizePhoneNumber(downlink chan string, uplink chan string, loglink chan string, trigger string, emission string, context map[string]string, slot string) {
	arabicPersianToEnglish := map[rune]rune{
		// Arabic numerals
		'٠': '0', '١': '1', '٢': '2', '٣': '3', '٤': '4',
		'٥': '5', '٦': '6', '٧': '7', '٨': '8', '٩': '9',
		// Persian numerals
		'۰': '0', '۱': '1', '۲': '2', '۳': '3', '۴': '4',
		'۵': '5', '۶': '6', '۷': '7', '۸': '8', '۹': '9',
	}
	for msg := range downlink {
		if msg == trigger {
			phone_number := context[slot]
			sanitizedPhoneNumber := strings.Builder{}

			for _, ch := range phone_number {
				if englishNum, exists := arabicPersianToEnglish[ch]; exists {
					sanitizedPhoneNumber.WriteRune(englishNum)
				} else if unicode.IsDigit(ch) {
					sanitizedPhoneNumber.WriteRune(ch)
				} else {
					loglink <- trigger + "->" + "ERROR: could not sanitize phone!"
					return
				}
			}
			context[slot] = sanitizedPhoneNumber.String()
			loglink <- trigger + "->" + emission
			uplink <- emission
		}
	}
}
func ValidatePhoneNumber(downlink chan string, uplink chan string, loglink chan string, trigger string, emission string, context map[string]string, slot string) {
	for msg := range downlink {
		if msg == trigger {
			phone_number := context[slot]
			if len(phone_number) == 11 && phone_number[:2] == "09" {
				loglink <- trigger + "->" + emission
				uplink <- emission
			} else {
				loglink <- trigger + "->" + "ERROR: could not validate phone!"
			}
		}
	}
}
func ValidateCode(downlink chan string, uplink chan string, loglink chan string, trigger string, emission string, context map[string]string, slot string) {
	for msg := range downlink {
		if msg == trigger {
			code := context[slot]
			if len(code) != 6 {
				return
			}
			for _, char := range code {
				if !unicode.IsDigit(char) || char < '0' || char > '9' {
					return
				}
			}
			loglink <- trigger + "->" + emission
			uplink <- emission
		}
	}
}
