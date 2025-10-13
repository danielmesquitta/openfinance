package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	awslambda "github.com/aws/aws-lambda-go/lambda"
	"github.com/danielmesquitta/openfinance/internal/app/lambda"
)

func main() {
	handler := lambda.NewLambdaHandler()

	awslambda.Start(func(ctx context.Context, event any) (any, error) {
		switch e := event.(type) {
		case events.APIGatewayProxyRequest:
			return handler.Handle(ctx, e)
		case events.CloudWatchEvent:
			err := handler.HandleScheduledEvent(ctx, e)
			if err != nil {
				log.Printf("Error handling scheduled event: %v", err)

				return nil, fmt.Errorf("error handling scheduled event: %w", err)
			}

			return map[string]string{"status": "success"}, nil
		default:
			err := handler.HandleScheduledEvent(ctx, events.CloudWatchEvent{})
			if err != nil {
				log.Printf("Error handling direct invocation: %v", err)

				return nil, fmt.Errorf("error handling direct invocation: %w", err)
			}

			return map[string]string{"status": "success"}, nil
		}
	})
}
