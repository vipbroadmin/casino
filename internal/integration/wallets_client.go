package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var ErrWalletConflict = errors.New("wallet already exists")

type WalletsClient struct {
	baseURL string
	client  *http.Client
}

type CreateWalletRequest struct {
	PlayerID string `json:"playerId"`
	Currency string `json:"currency"`
	Type     string `json:"type"`
}

func NewWalletsClient(baseURL string) *WalletsClient {
	return &WalletsClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *WalletsClient) CreateWallet(ctx context.Context, req CreateWalletRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/finances/wallets", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return ErrWalletConflict
	}
	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("wallets service returned %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	return nil
}
