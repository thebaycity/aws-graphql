package playground

import (
	_ "embed"
	"html/template"
	"net/http"
)

//go:embed template.gohtml
var tpl string
var page = template.Must(template.New("graphiql").Parse(tpl))

func Handler(title, endpoint, subscriptionEndpoint string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		err := page.Execute(w, map[string]interface{}{
			"title":                title,
			"endpoint":             endpoint,
			"subscriptionEndpoint": subscriptionEndpoint,
		})
		if err != nil {
			panic(err)
		}
	}
}
