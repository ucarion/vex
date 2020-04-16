package dynamo

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/ucarion/vex"
)

type Store struct {
	TableName string
	DB        *dynamodb.DynamoDB
}

func (s *Store) CreateFlag(ctx context.Context, flag vex.Flag) error {
	expr, err := json.Marshal(flag.Expression)
	if err != nil {
		return err
	}

	_, err = s.DB.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: &s.TableName,
		Item: map[string]*dynamodb.AttributeValue{
			"pk":   &dynamodb.AttributeValue{S: &flag.Namespace},
			"sk":   &dynamodb.AttributeValue{S: &flag.Name},
			"expr": &dynamodb.AttributeValue{B: expr},
		},
	})

	return err
}

func (s *Store) GetFlag(ctx context.Context, namespace, flagName string) (vex.Flag, error) {
	res, err := s.DB.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: &s.TableName,
		Key: map[string]*dynamodb.AttributeValue{
			"pk": &dynamodb.AttributeValue{S: &namespace},
			"sk": &dynamodb.AttributeValue{S: &flagName},
		},
	})

	if err != nil {
		return vex.Flag{}, err
	}

	var expr vex.FlagExpression
	if err := json.Unmarshal(res.Item["expr"].B, &expr); err != nil {
		return vex.Flag{}, err
	}

	return vex.Flag{Namespace: namespace, Name: flagName, Expression: expr}, nil
}
