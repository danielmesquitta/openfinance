package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	lambdaHandler "github.com/danielmesquitta/openfinance/internal/app/lambda"
)

func main() {
	lambda.Start(lambdaHandler.Handler)
}
