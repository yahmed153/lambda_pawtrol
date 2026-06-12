package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const TableName = "pawtrol"

type RequestBody struct {
	URL string `json:"url,omitempty"`
}

func buildResponse(httpStatusCode int, jsonBody string) events.APIGatewayV2HTTPResponse {
	return events.APIGatewayV2HTTPResponse{
		StatusCode: httpStatusCode,
		Body:       jsonBody,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}

func upsertItem(ctx context.Context, hostname string) error {
	key, err := attributevalue.MarshalMap(map[string]string{
		"hostname": hostname,
	})
	if err != nil {
		return err
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		return err
	}

	dynamoClient := dynamodb.NewFromConfig(cfg)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	expireAt := strconv.FormatInt(time.Now().AddDate(0, 0, 90).Unix(), 10)
	entryTimestamp := [1]string{timestamp}
	_, err = dynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:        new(TableName),
		Key:              key,
		UpdateExpression: new("SET #c = if_not_exists(#c, :zero) + :one, createdAt = if_not_exists(createdAt, :createdAt), updatedAt = :updatedAt, expireAt = :expireAt ADD entryTimestamps :entryTimestamp"),
		ExpressionAttributeNames: map[string]string{
			"#c": "count",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":zero":           &types.AttributeValueMemberN{Value: "0"},
			":one":            &types.AttributeValueMemberN{Value: "1"},
			":createdAt":      &types.AttributeValueMemberN{Value: timestamp},
			":updatedAt":      &types.AttributeValueMemberN{Value: timestamp},
			":expireAt":       &types.AttributeValueMemberN{Value: expireAt},
			":entryTimestamp": &types.AttributeValueMemberNS{Value: entryTimestamp[:]},
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

	// 2. Validate request body is a valid URL
	parsedURL, err := url.ParseRequestURI(body.URL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return buildResponse(http.StatusBadRequest,
			fmt.Sprintf(`{"error": "%s is not a valid URL"}`, body.URL)), nil
	}

	// 3. Store hostname in DB
	if err := upsertItem(ctx, parsedURL.Hostname()); err != nil {
		log.Println(err.Error())
		return buildResponse(http.StatusInternalServerError,
			`{"error": "Inserting hostname into database failed"}`), nil
	}

	// 4. Return successful response
	return buildResponse(http.StatusOK,
		fmt.Sprintf(`{"message": "Hostname: (%s) has been submitted. Thank you 😀 for helping us fight 🥊 malicious sites"}`, parsedURL.Hostname())), nil
}

func main() {
	lambda.Start(Handler)
}
