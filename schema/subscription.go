package schema

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"log"
	"time"
)

func Subscription() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Subscription",
		Fields: graphql.Fields{
			"messageCreated": &graphql.Field{
				Type: messageType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source, nil
				},
				Subscribe: func(p graphql.ResolveParams) (interface{}, error) {
					c := make(chan interface{})
					go func() {
						var i int

						for {
							i++
							feed := Message{ID: fmt.Sprintf("%d", i)}
							select {
							case <-p.Context.Done():
								log.Println("[RootSubscription] [Subscribe] ps canceled")
								close(c)
								return
							default:
								c <- feed
							}
							time.Sleep(250 * time.Millisecond)

							if i == 21 {
								close(c)
								return
							}
						}
					}()

					return c, nil
				},
			},
		},
		IsTypeOf:    nil,
		Description: "",
	})
}
