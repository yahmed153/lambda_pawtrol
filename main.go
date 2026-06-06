package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const TableName = "pawtrol"

type RequestBody struct {
	HostName string `json:"hostname,omitempty"`
}

func buildResponse(httpStatusCode int, jsonBody string) events.APIGatewayV2HTTPResponse {
	return events.APIGatewayV2HTTPResponse{
		StatusCode: httpStatusCode,
		Body:       jsonBody,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}

func insertItem(ctx context.Context, hostname string) error {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		return err
	}

	dynamoClient := dynamodb.NewFromConfig(cfg)
	_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: new(TableName),
		Item: map[string]types.AttributeValue{
			"hostname":   &types.AttributeValueMemberS{Value: hostname},
			"created_at": &types.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
		},
	})

	return err
}

func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// 1. Parse incoming JSON body
	var body RequestBody
	if req.Body == "" {
		return buildResponse(http.StatusBadRequest,
			`{"error": "Empty request body"}`), nil
	}
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		return buildResponse(http.StatusBadRequest,
			`{"error": "Invalid JSON request body"}`), nil
	}

	// 2. Validate hostname is a valid URL
	parsedURL, err := url.ParseRequestURI(body.HostName)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return buildResponse(http.StatusBadRequest,
			fmt.Sprintf(`{"error": "hostname is not a valid URL: %s"}`, body.HostName)), nil
	}

	// 3. Store hostname in DB
	if err := insertItem(ctx, parsedURL.Hostname()); err != nil {
		log.Println(err.Error())
		return buildResponse(http.StatusInternalServerError,
			fmt.Sprintf(`{"error": "Inserting hostname into database failed: %s"}`, err.Error())), nil
	}

	// 4. Return successful response
	return buildResponse(http.StatusInternalServerError,
		fmt.Sprintf(`{"message": "Thank you 😀 for helping us fight 🥊 malicious sites, hostname: (%s) has been submitted"}`, parsedURL.Hostname())), nil
}

func main() {
	// Start the Lambda runtime loop
	lambda.Start(Handler)
}
