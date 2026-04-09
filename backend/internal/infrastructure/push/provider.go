package push

import (
	"context"
	"encoding/json"
	"errors"
)

var ErrAllProvidersFailed = errors.New("all providers failed")

type Provider interface {
	Send(ctx context.Context, token string, title string, body string, data map[string]interface{}) error
	SendBatch(ctx context.Context, tokens []string, title string, body string, data map[string]interface{}) error
	Name() string
}

type InvalidTokenError struct {
	Token  string
	Reason string
}

func (e *InvalidTokenError) Error() string {
	return "invalid token: " + e.Reason
}

func IsInvalidToken(err error) bool {
	var invalidErr *InvalidTokenError
	return errors.As(err, &invalidErr)
}

func NewInvalidTokenError(token, reason string) *InvalidTokenError {
	return &InvalidTokenError{Token: token, Reason: reason}
}

func ParsePayload(payload json.RawMessage) (*NotificationPayload, error) {
	var p NotificationPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

type NotificationPayload struct {
	Title string                 `json:"title"`
	Body  string                 `json:"body"`
	Data  map[string]interface{} `json:"data"`
}

type TokenInfo struct {
	ID       string
	UserID   string
	Token    string
	Platform string
}

type MultiProvider struct {
	providers []Provider
}

func NewMultiProvider(providers ...Provider) *MultiProvider {
	return &MultiProvider{providers: providers}
}

func (m *MultiProvider) Send(ctx context.Context, token string, title string, body string, data map[string]interface{}) error {
	for _, provider := range m.providers {
		err := provider.Send(ctx, token, title, body, data)
		if err == nil {
			return nil
		}
		if IsInvalidToken(err) {
			return err
		}
	}
	return ErrAllProvidersFailed
}

func (m *MultiProvider) SendBatch(ctx context.Context, tokens []string, title string, body string, data map[string]interface{}) error {
	for _, provider := range m.providers {
		err := provider.SendBatch(ctx, tokens, title, body, data)
		if err == nil {
			return nil
		}
		if IsInvalidToken(err) {
			return err
		}
	}
	return ErrAllProvidersFailed
}

func (m *MultiProvider) Name() string {
	return "multi"
}
