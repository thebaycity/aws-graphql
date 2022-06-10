package ps

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
)

type PubSub struct {
	api *apigatewaymanagementapi.ApiGatewayManagementApi
}

func NewPubSub(awsWebsocketEndpoint, region string) *PubSub {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return &PubSub{api: apigatewaymanagementapi.New(sess, &aws.Config{
		Credentials: sess.Config.Credentials,
		Region:      aws.String(region),
		Endpoint:    aws.String(awsWebsocketEndpoint),
	})}
}

func (p *PubSub) PostToConnection(connectionId string, data interface{}) error {
	if data != nil {
		b, _ := json.Marshal(data)
		output, err := p.api.PostToConnection(&apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: aws.String(connectionId),
			Data:         b,
		})
		if err != nil {
			if output != nil {
				return fmt.Errorf("output: %s, error: %s", output.String(), err.Error())
			}
			return err
		}

	}
	return nil
}
