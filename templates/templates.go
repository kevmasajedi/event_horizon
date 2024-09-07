package templates

import (
	"html/template"
)

func GetTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"eq": func(a, b string) bool {
			return a == b
		},
	}
}
