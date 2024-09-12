package templates

import (
	"html/template"
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
