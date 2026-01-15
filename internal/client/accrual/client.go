package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
)

type AccrualStatus string

const (
	StatusInvalid   AccrualStatus = "INVALID"
	StatusProcessed AccrualStatus = "PROCESSED"
)

type AccrualResponse struct {
	Order   string        `json:"order"`
	Status  AccrualStatus `json:"status"`
	Accrual *float64      `json:"accrual,omitempty"`
}

type AccrualClient struct {
	baseURL    string
	httpClient *http.Client
	maxRetries int
	semaphore  chan struct{}
	logger     *zap.Logger
}

func NewAccrualClient(baseURL string, maxConcurrentRequests, maxRetries int, logger *zap.Logger) AccrualInterface {
	return &AccrualClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		maxRetries: maxRetries,
		semaphore:  make(chan struct{}, maxConcurrentRequests),
		logger:     logger,
	}
}

func (c *AccrualClient) GetAccrualInfo(ctx context.Context, orderNumber string) (*AccrualResponse, error) {
	requestURL, err := url.JoinPath(c.baseURL, "api/orders", orderNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		var result AccrualResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("invalid JSON response: %w", err)
		}
		return &result, nil

	case http.StatusNoContent:
		return nil, ErrOrderNotRegistered

	case http.StatusTooManyRequests:
		return nil, ErrTooManyRequests

	case http.StatusInternalServerError:
		return nil, ErrInternalServer

	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
