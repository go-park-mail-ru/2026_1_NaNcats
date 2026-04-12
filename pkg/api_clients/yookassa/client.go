package yookassa

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mailru/easyjson"
)

type Client struct {
	shopID    string
	secretKey string
	baseURL   string
	client    *http.Client
}

func NewClient(shopID, secretKey string) *Client {
	return &Client{
		shopID:    shopID,
		secretKey: secretKey,
		baseURL:   "https://api.yookassa.ru/v3",
		client: &http.Client{
			Timeout: time.Second * 15,
		},
	}
}

func (c *Client) CreatePayment(ctx context.Context, req CreatePaymentRequest) (*PaymentResponse, error) {
	data, err := easyjson.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal yookassa request: %w", err)
	}

	url := c.baseURL + "/payments"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotence-Key", uuid.New().String())
	httpReq.SetBasicAuth(c.shopID, c.secretKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request to yookassa failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("yookassa API returned error status: %d", resp.StatusCode)
	}

	var yookassaResponse PaymentResponse
	if err := easyjson.UnmarshalFromReader(resp.Body, &yookassaResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yookassa response: %w", err)
	}

	return &yookassaResponse, nil
}

func (c *Client) CreatePaymentMethod(ctx context.Context, req CreatePaymentMethodRequest) (*PaymentMethodResponse, error) {
	data, err := easyjson.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal yookassa request: %w", err)
	}

	url := c.baseURL + "/payment_methods"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotence-Key", uuid.New().String())
	httpReq.SetBasicAuth(c.shopID, c.secretKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request to yookassa failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("yookassa API returned error status: %d", resp.StatusCode)
	}

	var yookassaResponse PaymentMethodResponse
	if err := easyjson.UnmarshalFromReader(resp.Body, &yookassaResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yookassa response: %w", err)
	}

	return &yookassaResponse, nil
}
