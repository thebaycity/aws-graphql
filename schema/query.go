package schema

import "github.com/graphql-go/graphql"

type Message struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

var messageType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Message",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.ID,
		},
		"message": &graphql.Field{
			Type: graphql.String,
		},
	},
})

func Query() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"messages": &graphql.Field{
				Type: graphql.NewList(messageType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return []*Message{
						{
							ID:      "001",
							Message: "hi there",
						},
					}, nil
				},
				Description: "Messages",
			},
		},
		Description: "Query",
	})
}
