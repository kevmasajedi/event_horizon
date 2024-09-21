package templates

import (
	"encoding/json"
	"html/template"
	"math/rand"
	"strconv"
)

func GetTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"eq": func(a, b string) bool {
			return a == b
		},
		"unmarshal": func(data string) map[string]interface{} {
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(data), &result); err != nil {
				return nil
			}
			return result
		},
		"html": func(data string) template.HTML {
			return template.HTML(data)
		},
		"enumerator": func(base string, start byte, end byte) []string {
			var result []string
			for i := start; i <= end; i++ {
				result = append(result, base+strconv.Itoa(int(i)))
			}
			return result
		},
		"uid": func() string {
			const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			b := make([]byte, 8)
			for i := range b {
				b[i] = charset[rand.Intn(len(charset))]
			}
			return string(b)
		},
	}
}
