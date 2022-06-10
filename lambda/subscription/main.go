package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/graphql-go/graphql"
	"github.com/thebaycity/aws-graphql/logger"
	"github.com/thebaycity/aws-graphql/ps"
	"github.com/thebaycity/aws-graphql/schema"
	"go.uber.org/zap"
	"log"
	"net/http"
	"net/url"
	"os"
)

var pubSub *ps.PubSub

func handler(_ context.Context, req *events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	defer func() {
		_ = logger.Instance.Sync()
	}()
	if pubSub == nil {
		endpoint := url.URL{
			Scheme: "https",
			Host:   req.RequestContext.DomainName,
			Path:   req.RequestContext.Stage,
		}
		pubSub = ps.NewPubSub(endpoint.String(), os.Getenv("region"))
	}
	logger.Instance.Info("ps connect",
		zap.String("requestId", req.RequestContext.RequestID),
		zap.String("connectionId", req.RequestContext.ConnectionID), zap.Any("request", req))
	switch req.RequestContext.RouteKey {
	case "$connect":
		logger.Instance.Info("client connected", zap.String("connectionId", req.RequestContext.ConnectionID))
		break
	case "$disconnect":
		logger.Instance.Info("client disconnected", zap.String("connectionId", req.RequestContext.ConnectionID))
		break
	default:
		if req.Body != "" {
			var raw schema.SubscriptionRawParams
			if err := json.Unmarshal([]byte(req.Body), &raw); err != nil {
				break
			}
			if raw.Type == "start" {
				subscriptionCtx, subscriptionCancelFn := context.WithCancel(context.Background())
				subscribe(req.RequestContext.ConnectionID, subscriptionCtx, subscriptionCancelFn, raw)
			}
		} else {
			data := map[string]interface{}{
				"type":    "connection_ack",
				"message": req.Body,
			}
			err := pubSub.PostToConnection(req.RequestContext.ConnectionID, data)
			if err != nil {
				logger.Instance.Info("send ps client data error", zap.Any("error", err))
			}
		}
		break
	}
	return events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, nil
}

func subscribe(connectionId string, ctx context.Context, subscriptionCancelFn context.CancelFunc, p schema.SubscriptionRawParams) {
	go func() {
		s, _ := graphql.NewSchema(graphql.SchemaConfig{
			Query:        schema.Query(),
			Mutation:     nil,
			Subscription: schema.Subscription(),
		})
		subscribeChannel := graphql.Subscribe(graphql.Params{
			Schema:         s,
			RequestString:  p.Payload.Query,
			VariableValues: p.Payload.Variables,
			OperationName:  p.Payload.OperationName,
			Context:        ctx,
		})
		for {
			select {
			case <-ctx.Done():
				log.Printf("[SubscriptionsHandler] subscription ctx done")
				return
			case r, isOpen := <-subscribeChannel:
				if !isOpen {
					log.Printf("[SubscriptionsHandler] subscription channel closed")
					subscriptionCancelFn()
					return
				}
				if err := pubSub.PostToConnection(connectionId, map[string]interface{}{
					"type":    "data",
					"id":      p.ID,
					"payload": r.Data,
				}); err != nil {
					subscriptionCancelFn()
					return
				}
			}
		}
	}()
	logger.Instance.Info("done subscribe")
}

func main() {
	lambda.Start(handler)
}
