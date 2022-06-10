package schema

import (
	"encoding/json"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"io"
	"net/http"
	"strings"
)

type SubscriptionRawParams struct {
	ID      string    `json:"id"`
	Type    string    `json:"type"`
	Payload RawParams `json:"payload"`
}
type RawParams struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
	Extensions    map[string]interface{} `json:"extensions"`
}

func jsonDecode(r io.Reader, val interface{}) error {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	return dec.Decode(val)
}

func writeError(w http.ResponseWriter, message string) {
	json.NewEncoder(w).Encode(graphql.Result{
		Errors: []gqlerrors.FormattedError{
			{
				Message: message,
			},
		},
	})
}

func Handler() http.HandlerFunc {
	schema, _ := graphql.NewSchema(graphql.SchemaConfig{
		Query:        Query(),
		Mutation:     nil,
		Subscription: Subscription(),
	})
	return func(w http.ResponseWriter, r *http.Request) {
		var params RawParams
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			params = RawParams{
				Query:         r.URL.Query().Get("query"),
				OperationName: r.URL.Query().Get("operationName"),
				Variables:     nil,
				Extensions:    nil,
			}
			if variables := r.URL.Query().Get("variables"); variables != "" {
				if err := jsonDecode(strings.NewReader(variables), &params.Variables); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					writeError(w, "variables could not be decoded")
					return
				}
			}
		}
		if r.Method == http.MethodPost {
			if err := jsonDecode(r.Body, &params); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				writeError(w, err.Error())
				return
			}
		}
		result := graphql.Do(graphql.Params{
			Schema:         schema,
			RequestString:  params.Query,
			OperationName:  params.OperationName,
			VariableValues: params.Variables,
		})
		if len(result.Errors) > 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
		}
		json.NewEncoder(w).Encode(result)
	}
}
