package client

import (
	"context"
	"encoding/json"
	"fmt"
	"kate/shared/httpx"
	"net/http"
	"time"
)

type AuthClient struct {
	httpClient *httpx.Client
}

func NewAuthClient(baseURL string, timeout time.Duration) *AuthClient {
	return &AuthClient{
		httpClient: httpx.NewClient(baseURL, timeout),
	}
}

type verifyResponse struct {
	Valid   bool   `json:"valid"`
	Subject string `json:"subject"`
	Error   string `json:"error"`
}

func (c *AuthClient) VerifyToken(ctx context.Context, token string) (bool, error) {
	url := fmt.Sprintf("%s/v1/auth/verify", c.httpClient.BaseURL())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.DoWithRequestID(ctx, req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var vresp verifyResponse
		if err := json.NewDecoder(resp.Body).Decode(&vresp); err != nil {
			return false, err
		}
		return vresp.Valid, nil
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return false, nil
	}
	return false, fmt.Errorf("auth service returned status %d", resp.StatusCode)
}
