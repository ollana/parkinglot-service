package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const (
	// CostPerHour defines the charge for parking per hour.
	CostPerHour float64 = 10
	// ParkingDurationUnit in minutes defines the increments for charging.
	ParkingDurationUnit time.Duration = 15 * time.Minute
)

var dbClient dynamoDBClientInterface

// TicketId represents the ticket ID returned on entry.
type TicketId struct {
	ID string `json:"ticketId"`
}

// ExitDetails represents the details returned on exit.
type ExitDetails struct {
	License    string  `json:"license"`
	ParkedTime string  `json:"parkedTime"`
	ParkingLot int     `json:"parkingLot"`
	Charge     float64 `json:"charge"`
}

// entryHandler generates a new ticket for a car entering the parking lot.
func entryHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get the query parameters.
	license := request.QueryStringParameters["plate"]
	parkingLotIDStr := request.QueryStringParameters["parkingLot"]
	parkingLotID, err := strconv.Atoi(parkingLotIDStr)

	if err != nil {
		return apiResponse(http.StatusBadRequest, "Invalid parking lot ID")
	}

	// Store the ticket in the database
	ticket := dbTicket{
		TicketID:   uuid.New().String(), // Generate a new random UUID for the ticket ID.
		License:    license,
		ParkingLot: parkingLotID,
		EntryTime:  time.Now().Format(time.RFC3339),
		Closed:     false,
	}
	err = dbClient.StoreTicket(ctx, ticket)
	if err != nil {
		fmt.Println(fmt.Errorf("error storing ticket: %v", err))
		return apiResponse(http.StatusInternalServerError, "Failed to store ticket")
	}

	// Write the ticket ID back to the client.
	resp := TicketId{
		ID: ticket.TicketID,
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		fmt.Println(fmt.Errorf("error on format response: %v", err))
		return apiResponse(http.StatusInternalServerError, "Failed to format response")
	}

	fmt.Println("New ticket generated:", ticket.TicketID)
	return apiResponse(http.StatusOK, string(jsonResp))

}

// exitHandler calculates the parking charge and total parked time.
func exitHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	ticketID := request.QueryStringParameters["ticketId"]

	// Get the ticket from the database
	ticket, err := dbClient.GetTicket(ctx, ticketID)
	if err != nil {
		fmt.Println(fmt.Errorf("error getting ticket %s: %v", ticketID, err))
		return apiResponse(http.StatusInternalServerError, "Failed to get ticket")
	}

	if ticket == nil {
		return apiResponse(http.StatusNotFound, "Ticket ID not found")
	}

	// if ticket is not closed, calculate the charge and close it
	if !ticket.Closed {
		// calculate parked time
		entryTime, err := time.Parse(time.RFC3339, ticket.EntryTime)
		if err != nil {
			fmt.Println(fmt.Errorf("error parsing entry time: %v", err))
			return apiResponse(http.StatusInternalServerError, "Failed to parse entry time")
		}

		parkedDuration := time.Since(entryTime)

		// Calculate charge based on the parked duration. Round up to the nearest 15 minutes.
		increments := calculateSegments(parkedDuration)
		charge := float64(increments) * (CostPerHour / 4)

		// Update the ticket in the database
		ticket.ParkedTime = parkedDuration
		ticket.Charge = charge
		ticket.Closed = true
		err = dbClient.updateTicket(ctx, *ticket)
		if err != nil {
			fmt.Println(fmt.Errorf("error updating ticket: %v", err))
			return apiResponse(http.StatusInternalServerError, "Failed to update ticket exit")
		}
	}
	// return the ticket exit details
	details := ExitDetails{
		License:    ticket.License,
		ParkedTime: ticket.ParkedTime.Round(time.Second).String(),
		ParkingLot: ticket.ParkingLot,
		Charge:     ticket.Charge,
	}

	jsonResp, err := json.Marshal(details)
	if err != nil {
		fmt.Println(fmt.Errorf("error formatting exit details: %v", err))
		return apiResponse(http.StatusInternalServerError, "Failed to format exit details")
	}

	fmt.Println(fmt.Sprintf("Exit ticket details generated: %+v", details))
	return apiResponse(http.StatusOK, string(jsonResp))

}

func calculateSegments(duration time.Duration) int {
	// Add the segment duration minus one nanosecond before dividing, to ensure rounding up.
	segments := (duration + ParkingDurationUnit - time.Nanosecond) / ParkingDurationUnit
	return int(segments)
}

// Lambda's API Gateway proxy request handler to replace the standard main function
func HandleLambdaEvent(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch request.HTTPMethod {
	case http.MethodPost:
		switch request.Path {
		case "/entry":
			return entryHandler(ctx, request)
		case "/exit":
			return exitHandler(ctx, request)
		default:
			return apiResponse(http.StatusNotFound, "Not Found")
		}
	default:
		return apiResponse(http.StatusMethodNotAllowed, "Method Not Allowed")
	}
}

// Helper function to generate API Gateway Proxy response
func apiResponse(statusCode int, body string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       body,
	}, nil
}

func main() {
	var err error
	dbClient, err = NewDynamoDBClient()
	if err != nil {
		panic(fmt.Errorf("Error loading configuration:" + err.Error()))
	}

	lambda.Start(HandleLambdaEvent)

}
