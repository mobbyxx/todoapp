package push

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const fcmEndpoint = "https://fcm.googleapis.com/fcm/send"

type FCMProvider struct {
	apiKey     string
	httpClient *http.Client
}

func NewFCMProvider(apiKey string) *FCMProvider {
	return &FCMProvider{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (f *FCMProvider) Name() string {
	return "fcm"
}

func (f *FCMProvider) Send(ctx context.Context, token string, title string, body string, data map[string]interface{}) error {
	if f.apiKey == "" {
		return fmt.Errorf("fcm api key not configured")
	}

	payload := fcmPayload{
		To: token,
		Notification: fcmNotification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal fcm payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fcmEndpoint, strings.NewReader(string(jsonBody)))
	if err != nil {
		return fmt.Errorf("failed to create fcm request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "key="+f.apiKey)

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fcm request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fcm returned status %d", resp.StatusCode)
	}

	var result fcmResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode fcm response: %w", err)
	}

	if result.Failure > 0 && len(result.Results) > 0 {
		errResult := result.Results[0]
		if errResult.Error == "InvalidRegistration" || errResult.Error == "NotRegistered" {
			return NewInvalidTokenError(token, errResult.Error)
		}
		if errResult.Error != "" {
			return fmt.Errorf("fcm error: %s", errResult.Error)
		}
	}

	return nil
}

func (f *FCMProvider) SendBatch(ctx context.Context, tokens []string, title string, body string, data map[string]interface{}) error {
	if len(tokens) == 0 {
		return nil
	}

	if f.apiKey == "" {
		return fmt.Errorf("fcm api key not configured")
	}

	var lastErr error
	for _, token := range tokens {
		if err := f.Send(ctx, token, title, body, data); err != nil {
			lastErr = err
			if IsInvalidToken(err) {
				return err
			}
		}
	}

	return lastErr
}

type fcmPayload struct {
	To           string                 `json:"to"`
	Notification fcmNotification        `json:"notification"`
	Data         map[string]interface{} `json:"data,omitempty"`
}

type fcmNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type fcmResponse struct {
	MulticastID  int64           `json:"multicast_id"`
	Success      int             `json:"success"`
	Failure      int             `json:"failure"`
	CanonicalIDs int             `json:"canonical_ids"`
	Results      []fcmResultItem `json:"results"`
}

type fcmResultItem struct {
	MessageID      string `json:"message_id,omitempty"`
	RegistrationID string `json:"registration_id,omitempty"`
	Error          string `json:"error,omitempty"`
}
