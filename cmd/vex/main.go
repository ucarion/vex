package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/segmentio/cli"
	"github.com/ucarion/vex"
	"github.com/ucarion/vex/dynamo"
)

func main() {
	type config struct {
		DynamoDBEndpoint  string `flag:"--dynamodb-endpoint" help:"DynamoDB endpoint URL" default:"http://localhost:8000"`
		DynamoDBTableName string `flag:"--dynamodb-table-name" help:"DynamoDB table name" default:"vex"`
	}

	cli.Exec(cli.CommandSet{
		"flags": cli.CommandSet{
			"create": cli.Command(func(ctx context.Context, config config, namespace, flagName, expression string) error {
				session := session.New(&aws.Config{Endpoint: &config.DynamoDBEndpoint})
				dynamo := dynamo.Store{
					TableName: config.DynamoDBTableName,
					DB:        dynamodb.New(session),
				}

				var expr vex.FlagExpression
				if err := json.Unmarshal([]byte(expression), &expr); err != nil {
					return err
				}

				return dynamo.CreateFlag(ctx, vex.Flag{
					Namespace:  namespace,
					Name:       flagName,
					Expression: expr,
				})
			}),

			"get": cli.Command(func(ctx context.Context, config config, namespace, flagName string) error {
				session := session.New(&aws.Config{Endpoint: &config.DynamoDBEndpoint})
				dynamo := dynamo.Store{
					TableName: config.DynamoDBTableName,
					DB:        dynamodb.New(session),
				}

				flag, err := dynamo.GetFlag(ctx, namespace, flagName)
				if err != nil {
					return err
				}

				expr, err := json.Marshal(flag.Expression)
				if err != nil {
					return err
				}

				fmt.Println(string(expr))
				return nil
			}),

			"on": cli.Command(func(ctx context.Context, config config, namespace, flagName, value string) error {
				session := session.New(&aws.Config{Endpoint: &config.DynamoDBEndpoint})
				vex.DefaultClient = &dynamo.Store{
					TableName: config.DynamoDBTableName,
					DB:        dynamodb.New(session),
				}

				on, err := vex.On(ctx, namespace, flagName, value)
				if err != nil {
					return err
				}

				fmt.Println(on)
				return nil
			}),
		},
	})
}
