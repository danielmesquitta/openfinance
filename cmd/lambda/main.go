package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	applambda "github.com/danielmesquitta/openfinance/internal/app/lambda"
)

func main() {
	lambda.Start(applambda.Handler)
}
