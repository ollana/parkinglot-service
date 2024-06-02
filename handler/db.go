package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"time"
)

type dbTicket struct {
	TicketID   string        `json:"TicketID"`
	License    string        `json:"License"`
	ParkingLot int           `json:"ParkingLot"`
	EntryTime  string        `json:"EntryTime"`
	ParkedTime time.Duration `json:"ParkedTime"`
	Charge     float64       `json:"Charge"`
	Closed     bool          `json:"Closed"`
}

type dynamoDBClientInterface interface {
	StoreTicket(ctx context.Context, ticket dbTicket) error
	GetTicket(ctx context.Context, ticketId string) (*dbTicket, error)

	updateTicket(ctx context.Context, ticket dbTicket) error
}

const (
	TableName  = "tickets"
	PrimaryKey = "TicketID"
)

type dynamoDBClient struct {
	client *dynamodb.Client
}

func NewDynamoDBClient() (dynamoDBClientInterface, error) {
	dynamoClient := &dynamoDBClient{}
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-west-2"))
	if err != nil {
		fmt.Println("Error loading configuration, ", err)
		return nil, err
	}
	dbClient := dynamodb.NewFromConfig(cfg)
	dynamoClient.client = dbClient

	return dynamoClient, nil
}

func (d *dynamoDBClient) StoreTicket(ctx context.Context, ticket dbTicket) error {

	// Serialize the Ticket into a map[string]AttributeValue
	av, err := attributevalue.MarshalMap(ticket)
	if err != nil {
		return err
	}

	// Create PutItem input
	input := &dynamodb.PutItemInput{
		TableName: aws.String(TableName),
		Item:      av,
	}

	// Write to DynamoDB
	_, err = d.client.PutItem(ctx, input)
	if err != nil {
		return err
	}
	return nil

}

func (d *dynamoDBClient) GetTicket(ctx context.Context, ticketId string) (*dbTicket, error) {
	// Create GetItem input
	id, err := attributevalue.Marshal(ticketId)
	if err != nil {
		return nil, err
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(TableName),
		Key:       map[string]types.AttributeValue{PrimaryKey: id},
	}

	// Get item from DynamoDB
	result, err := d.client.GetItem(ctx, input)
	if err != nil {
		return nil, err
	}

	// If result.Item is empty, no item with the provided ID exists
	if result.Item == nil {
		return nil, nil
	}

	// Unmarshal the result
	var ticket dbTicket
	err = attributevalue.UnmarshalMap(result.Item, &ticket)
	if err != nil {
		return nil, err
	}

	return &ticket, nil

}

func (d *dynamoDBClient) updateTicket(ctx context.Context, ticket dbTicket) error {
	av, err := attributevalue.MarshalMap(ticket)
	if err != nil {
		return err
	}

	// Create PutItem input
	input := &dynamodb.PutItemInput{
		TableName: aws.String(TableName),
		Item:      av,
	}

	// Write to DynamoDB
	_, err = d.client.PutItem(ctx, input)
	if err != nil {
		return err
	}
	return nil
}
