package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	app "github.com/danielmesquitta/openfinance/internal/app/lambda"
)

func main() {
	handler, err := app.NewLambdaHandler()
	if err != nil {
		log.Fatalf("failed to initialize lambda handler: %v", err)
	}

	lambda.Start(func(ctx context.Context) (any, error) {
		return handler.Handle(ctx)
	})
}
