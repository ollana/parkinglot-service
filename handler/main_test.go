package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

// mock the dynamoDBClientInterface
type mockDBClient struct {
	tickets map[string]dbTicket
}

func NewMockDBClient() dynamoDBClientInterface {
	return &mockDBClient{
		tickets: make(map[string]dbTicket),
	}
}

func (m *mockDBClient) StoreTicket(ctx context.Context, ticket dbTicket) error {
	m.tickets[ticket.TicketID] = ticket
	return nil

}
func (m *mockDBClient) GetTicket(ctx context.Context, ticketId string) (*dbTicket, error) {
	if ticket, ok := m.tickets[ticketId]; ok {
		return &ticket, nil
	}
	return nil, nil
}

func (m *mockDBClient) updateTicket(ctx context.Context, ticket dbTicket) error {
	m.tickets[ticket.TicketID] = ticket
	return nil
}

func TestCalculateSegments(t *testing.T) {
	assert.Equal(t, 0, calculateSegments(0*time.Minute))
	assert.Equal(t, 1, calculateSegments(1*time.Minute))
	assert.Equal(t, 1, calculateSegments(14*time.Minute))
	assert.Equal(t, 1, calculateSegments(15*time.Minute))
	assert.Equal(t, 2, calculateSegments(16*time.Minute))
	assert.Equal(t, 2, calculateSegments(30*time.Minute))
	assert.Equal(t, 3, calculateSegments(45*time.Minute))
	assert.Equal(t, 4, calculateSegments(60*time.Minute))
	assert.Equal(t, 5, calculateSegments(75*time.Minute))
	assert.Equal(t, 6, calculateSegments(76*time.Minute))
}

// //   "rawQueryString": "plate=123&parkingLot=1",
func TestEntryHandler(t *testing.T) {
	c := context.Background()
	dbClient = NewMockDBClient()

	// test case for happy path
	t.Run("happy path", func(t *testing.T) {
		request := events.APIGatewayProxyRequest{
			QueryStringParameters: map[string]string{
				"plate":      "123",
				"parkingLot": "1",
			},
		}
		res, err := entryHandler(c, request)
		// check the response
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		// check the response body
		// convert the response body response object and validate
		var ticket TicketId
		err = json.Unmarshal([]byte(res.Body), &ticket)
		assert.Nil(t, err)
		assert.NotNil(t, ticket.ID)
	})

	// test case for invalid parking lot ID
	t.Run("invalid parking lot ID", func(t *testing.T) {
		request := events.APIGatewayProxyRequest{
			QueryStringParameters: map[string]string{
				"plate":      "123",
				"parkingLot": "invalid",
			},
		}
		res, err := entryHandler(c, request)
		// check the response
		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		assert.Contains(t, res.Body, "Invalid parking lot ID")

	})

}

func TestExitHandler(t *testing.T) {
	c := context.Background()
	dbClient = NewMockDBClient()

	// test case for happy path
	t.Run("happy path", func(t *testing.T) {
		request := events.APIGatewayProxyRequest{
			QueryStringParameters: map[string]string{
				"plate":      "123",
				"parkingLot": "1",
			},
		}
		res, err := entryHandler(c, request)
		// check the response
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		var ticket TicketId
		err = json.Unmarshal([]byte(res.Body), &ticket)
		assert.Nil(t, err)

		// create a new exit request for ticket
		request = events.APIGatewayProxyRequest{
			QueryStringParameters: map[string]string{
				"ticketId": fmt.Sprint(ticket.ID),
			},
		}

		time.Sleep(1 * time.Second)

		res, err = exitHandler(c, request)
		// check the response
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		// check the response body
		// convert the response body response object and validate
		var exitDetails ExitDetails
		err = json.Unmarshal([]byte(res.Body), &exitDetails)
		assert.Nil(t, err)

		duration, err := time.ParseDuration(exitDetails.ParkedTime)
		assert.Nil(t, err)
		assert.Equal(t, "123", exitDetails.License)
		assert.Equal(t, 1, exitDetails.ParkingLot)
		assert.Greater(t, duration, 0*time.Millisecond)
	})

	// test case for ticket not found
	t.Run("ticket not found", func(t *testing.T) {
		request := events.APIGatewayProxyRequest{
			QueryStringParameters: map[string]string{
				"ticketId": "1000",
			},
		}
		res, err := exitHandler(c, request)
		assert.Nil(t, err)

		// check the response
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
		// check the response body
		assert.Contains(t, res.Body, "not found")
	})

}
