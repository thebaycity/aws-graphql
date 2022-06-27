package ps

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
)

var (
	sess *session.Session
)

type PubSub struct {
	api *apigatewaymanagementapi.ApiGatewayManagementApi
}

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
}

func NewPubSub(awsWebsocketEndpoint, region string) *PubSub {
	return &PubSub{api: apigatewaymanagementapi.New(sess, &aws.Config{
		Credentials: sess.Config.Credentials,
		Region:      aws.String(region),
		Endpoint:    aws.String(awsWebsocketEndpoint),
	})}
}

func (p *PubSub) PostToConnection(connectionId string, data interface{}) error {
	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			p.PostToConnection(connectionId, map[string]interface{}{
				"type": "error",
				"payload": map[string]interface{}{
					"errors": []string{fmt.Sprintf("encode json error: %s", err.Error())},
				},
			})
		}
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
