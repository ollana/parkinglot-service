package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
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

// Ticket represents a parking ticket with entry details.
type Ticket struct {
	ID         int       `json:"id"`
	License    string    `json:"license"`
	ParkingLot int       `json:"parkingLot"`
	EntryTime  time.Time `json:"entryTime"`
}

type TicketId struct {
	ID int `json:"ticketId"`
}

// ExitDetails represents the details returned on exit.
type ExitDetails struct {
	License    string        `json:"license"`
	ParkedTime time.Duration `json:"parkedTime"`
	ParkingLot int           `json:"parkingLot"`
	Charge     float64       `json:"charge"`
}

var (
	// TODO: Can use database for persistence.
	tickets   = make(map[int]Ticket)
	ticketMux sync.Mutex
	nextID    = 1
)

// entryHandler generates a new ticket for a car entering the parking lot.
func entryHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get the query parameters.
	license := request.QueryStringParameters["plate"]
	parkingLotIDStr := request.QueryStringParameters["parkingLot"]
	parkingLotID, err := strconv.Atoi(parkingLotIDStr)

	if err != nil {
		return apiResponse(http.StatusBadRequest, "Invalid parking lot ID")
	}

	// Create and store the parking ticket.
	ticketMux.Lock()
	defer ticketMux.Unlock()

	ticket := Ticket{
		ID:         nextID,
		License:    license,
		ParkingLot: parkingLotID,
		EntryTime:  time.Now(),
	}

	tickets[nextID] = ticket
	nextID++

	// Write the ticket ID back to the client.
	resp := TicketId{
		ID: ticket.ID,
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		return apiResponse(http.StatusInternalServerError, err.Error())
	}

	fmt.Println("New ticket generated:", ticket.ID)
	return apiResponse(http.StatusOK, string(jsonResp))

}

// exitHandler calculates the parking charge and total parked time.
func exitHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	ticketIDStr := request.QueryStringParameters["ticketId"]
	ticketID, err := strconv.Atoi(ticketIDStr)

	if err != nil {
		return apiResponse(http.StatusBadRequest, "Invalid ticket ID")
	}

	ticketMux.Lock()
	ticket, ok := tickets[ticketID]
	if !ok {
		ticketMux.Unlock()
		return apiResponse(http.StatusNotFound, "Ticket ID not found")

	}
	delete(tickets, ticketID)
	ticketMux.Unlock()

	// Calculate the total parked time.
	exitTime := time.Now()
	parkedDuration := exitTime.Sub(ticket.EntryTime)

	// Calculate charge based on the parked duration. Round up to the nearest 15 minutes.
	increments := calculateSegments(parkedDuration)
	charge := float64(increments) * (CostPerHour / 4)

	// return park duration in minutes

	details := ExitDetails{
		License:    ticket.License,
		ParkedTime: parkedDuration,
		ParkingLot: ticket.ParkingLot,
		Charge:     charge,
	}

	jsonResp, err := json.Marshal(details)
	if err != nil {
		return apiResponse(http.StatusInternalServerError, err.Error())
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

	lambda.Start(HandleLambdaEvent)

}
