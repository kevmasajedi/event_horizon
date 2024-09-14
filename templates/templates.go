package templates

import (
	"encoding/json"
	"html/template"
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
	}
}
