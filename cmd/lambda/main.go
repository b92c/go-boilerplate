package main

import (
	"github.com/b92c/go-boilerplate/internal/handler"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(handler.LambdaHandler)
}
