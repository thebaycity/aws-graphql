package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/thebaycity/aws-graphql/logger"
	"github.com/thebaycity/aws-graphql/ps"
	"github.com/thebaycity/aws-graphql/schema"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"os"
)

var (
	pubSub *ps.PubSub
)

func reply(id string, data interface{}) {
	err := pubSub.PostToConnection(id, data)
	if err != nil {
		logger.Instance.Info("send ps client data error", zap.Any("error", err))
	}
}

func handler(_ context.Context, req *events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	defer func() {
		_ = logger.Instance.Sync()
	}()
	var body = `{"type":"ka"}`
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
		zap.String("routerKey", req.RequestContext.RouteKey),
		zap.String("connectionId", req.RequestContext.ConnectionID), zap.Any("request", req))
	switch req.RequestContext.RouteKey {
	case "$connect":
		body = `{"type": "connection_ack","payload":{"connectionTimeoutMs":300000}}`
		break
	case "$disconnect":
		break
	default:
		logger.Instance.Info("client subscribed", zap.String("connectionId", req.RequestContext.ConnectionID), zap.String("body", req.Body))
		if req.Body == "" {
			break
		}
		var raw schema.SubscriptionRawParams
		if err := json.Unmarshal([]byte(req.Body), &raw); err != nil {
			break
		}
		if raw.Type == "ping" {
			body = `{"type":"pong"}`
		}
		if raw.Type == "connection_init" {
			body = `{"type": "connection_ack","payload":{"connectionTimeoutMs":300000}}`
		}
		if raw.Type == "subscribe" {
			body = fmt.Sprintf(`{"type":"next","id":"000","payload":{}}`)
		}
		if raw.Type == "complete" {
			// todo remove subscription id
		}
		break
	}
	return events.APIGatewayProxyResponse{Headers: map[string]string{
		"Sec-WebSocket-Protocol": "graphql-transport-ws",
	}, StatusCode: http.StatusOK, Body: body}, nil
}

func main() {
	lambda.Start(handler)
}
