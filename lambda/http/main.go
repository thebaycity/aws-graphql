package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/gorilla/mux"
	"github.com/thebaycity/aws-graphql/logger"
	"github.com/thebaycity/aws-graphql/playground"
	"github.com/thebaycity/aws-graphql/schema"
	"net/http"
	"os"
)

var router *mux.Router

func env(key, value string) string {
	v := os.Getenv(key)
	if v == "" {
		return value
	}
	return v
}
func init() {
	endpoint := env("endpoint", "http://localhost:8080/")
	router = mux.NewRouter()
	router.Handle("/", playground.Handler("Playground", endpoint+"graphql", os.Getenv("subscription_endpoint")))
	router.Handle("/graphql", schema.Handler())
}
func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	defer func() {
		_ = logger.Instance.Sync()
	}()
	return httpadapter.New(router).ProxyWithContext(ctx, req)
}
func main() {
	if os.Getenv("exec_env") == "lambda" {
		lambda.Start(handler)
	} else {
		http.ListenAndServe(":8080", router)
	}
}
