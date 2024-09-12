package templates

import (
	"html/template"
	"os"
)

func GetTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"eq": func(a, b string) bool {
			return a == b
		},
		"base_url": func() string {
			return os.Getenv("DISPATCH_SUB_DOMAIN")
		},
	}
}
