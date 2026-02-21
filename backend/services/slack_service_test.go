package services

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHTTPClient is a mock implementation of HTTPClient
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestSlackService_SendNotification(t *testing.T) {
	mockClient := new(MockHTTPClient)
	service := &SlackService{
		WebhookURL: "https://example.com/webhook",
		Client:     mockClient,
	}

	t.Run("Success via Webhook", func(t *testing.T) {
		mockClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString("ok")),
		}, nil).Once()

		err := service.SendNotification("", "Hello Slack", nil)
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("Failure - Network Error", func(t *testing.T) {
		mockClient.On("Do", mock.AnythingOfType("*http.Request")).Return(nil, errors.New("network error")).Once()

		err := service.SendNotification("", "Hello Slack", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send slack request")
		mockClient.AssertExpectations(t)
	})

	t.Run("Failure - API Error", func(t *testing.T) {
		mockClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&http.Response{
			StatusCode: 500,
			Body:       io.NopCloser(bytes.NewBufferString("server error")),
		}, nil).Once()

		err := service.SendNotification("", "Hello Slack", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "status code 500")
		mockClient.AssertExpectations(t)
	})
}

func TestSlackService_SendAlert(t *testing.T) {
	mockClient := new(MockHTTPClient)
	service := &SlackService{
		WebhookURL: "https://example.com/webhook/test",
		Client:     mockClient,
	}

	t.Run("Send Critical Alert", func(t *testing.T) {
		mockClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString("ok")),
		}, nil).Once()

		err := service.SendAlert("general", "System Down", "Critical failure", "critical")
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})
}
