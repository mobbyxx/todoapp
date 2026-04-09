package push

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const expoEndpoint = "https://exp.host/--/api/v2/push/send"

type ExpoProvider struct {
	httpClient *http.Client
}

func NewExpoProvider() *ExpoProvider {
	return &ExpoProvider{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (e *ExpoProvider) Name() string {
	return "expo"
}

func (e *ExpoProvider) Send(ctx context.Context, token string, title string, body string, data map[string]interface{}) error {
	if !IsExpoPushToken(token) {
		return fmt.Errorf("invalid expo push token format")
	}

	message := expoMessage{
		To:       token,
		Title:    title,
		Body:     body,
		Data:     data,
		Sound:    "default",
		Priority: "high",
	}

	return e.sendMessages(ctx, []expoMessage{message})
}

func (e *ExpoProvider) SendBatch(ctx context.Context, tokens []string, title string, body string, data map[string]interface{}) error {
	if len(tokens) == 0 {
		return nil
	}

	messages := make([]expoMessage, 0, len(tokens))
	for _, token := range tokens {
		if !IsExpoPushToken(token) {
			continue
		}
		messages = append(messages, expoMessage{
			To:       token,
			Title:    title,
			Body:     body,
			Data:     data,
			Sound:    "default",
			Priority: "high",
		})
	}

	if len(messages) == 0 {
		return fmt.Errorf("no valid expo push tokens")
	}

	return e.sendMessages(ctx, messages)
}

func (e *ExpoProvider) sendMessages(ctx context.Context, messages []expoMessage) error {
	jsonBody, err := json.Marshal(messages)
	if err != nil {
		return fmt.Errorf("failed to marshal expo payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, expoEndpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create expo request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip, deflate")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("expo request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expo returned status %d", resp.StatusCode)
	}

	var result expoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode expo response: %w", err)
	}

	if result.Errors != nil && len(result.Errors) > 0 {
		return fmt.Errorf("expo api error: %s", result.Errors[0].Message)
	}

	for i, ticket := range result.Data {
		if ticket.Status == "error" {
			if ticket.Details != nil && ticket.Details.Error == "DeviceNotRegistered" {
				return NewInvalidTokenError(messages[i].To, ticket.Message)
			}
			return fmt.Errorf("expo push error: %s", ticket.Message)
		}
	}

	return nil
}

func IsExpoPushToken(token string) bool {
	return len(token) > 0 && (token[:13] == "ExponentPush[" || token[:8] == "ExpoPush")
}

type expoMessage struct {
	To       string                 `json:"to"`
	Title    string                 `json:"title,omitempty"`
	Body     string                 `json:"body"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Sound    string                 `json:"sound,omitempty"`
	Priority string                 `json:"priority,omitempty"`
}

type expoResponse struct {
	Data   []expoTicket `json:"data"`
	Errors []expoError  `json:"errors,omitempty"`
}

type expoTicket struct {
	ID       string         `json:"id"`
	Status   string         `json:"status"`
	Message  string         `json:"message,omitempty"`
	Details  *expoDetails   `json:"details,omitempty"`
}

type expoDetails struct {
	Error string `json:"error,omitempty"`
}

type expoError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
