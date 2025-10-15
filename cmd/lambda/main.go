package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	app "github.com/danielmesquitta/openfinance/internal/app/lambda"
)

func main() {
	handler := app.NewLambdaHandler()

	lambda.Start(func(ctx context.Context) (any, error) {
		return handler.Handle(ctx)
	})
}
