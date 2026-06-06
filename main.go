package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// RequestBody defines the expected JSON input
type RequestBody struct {
	HostName string `json:"hostname,omitempty"`
}

// ResponseBody defines the JSON output
type ResponseBody struct {
	Message string `json:"message,omitempty"`
}

// Handler processes the API Gateway HTTP request
func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// 1. Parse incoming JSON body
	var body RequestBody
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusBadRequest,
			Body:       `{"error": "Invalid JSON request body"}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}

	// 2. Validate hostname is a valid URL
	parsedURL, err := url.ParseRequestURI(body.HostName)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusBadRequest,
			Body:       fmt.Sprintf(`{"error": "hostname is not a valid URL: %s"}`, body.HostName),
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}

	// 3. Construct JSON response
	resBytes, err := json.Marshal(ResponseBody{Message: fmt.Sprintf("Hello, %s! Your Lambda function worked perfectly.", body.HostName)})
	if err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       `{"error": "Internal server error"}`,
		}, nil
	}

	// 4. Return successful proxy response
	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(resBytes),
	}, nil
}

func main() {
	// Start the Lambda runtime loop
	lambda.Start(Handler)
}
