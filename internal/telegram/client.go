package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	token string
}

func NewClient(token string) *Client {
	return &Client{
		token: token,
	}
}

func (c *Client) EditMessage(
	ctx context.Context,
	chatID int64,
	messageID int,
	text string,
	eventID int64,
	stats map[string]int,
) error {
	url := fmt.Sprintf(
		"https://api.telegram.org/bot%s/editMessageText",
		c.token,
	)

	reqData := map[string]any{
		"chat_id":      chatID,
		"message_id":   messageID,
		"text":         text,
		"reply_markup": InlineKeyboard(eventID, stats),
	}

	body, err := json.Marshal(reqData)
	if err != nil {
		return fmt.Errorf("edit message marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("edit message create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	//nolint:gosec
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("edit message send: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	return nil
}

func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	req := map[string]any{
		"chat_id": chatID,
		"text":    text,
	}

	return c.post(ctx, "sendMessage", req, nil)
}

type SendMessageResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		MessageID int64 `json:"message_id"`
	} `json:"result"`
}

func (c *Client) SendEventMessage(
	ctx context.Context,
	chatID int64,
	text string,
	eventID int64,
	stats map[string]int,
) (*SendMessageResponse, error) {
	req := map[string]any{
		"chat_id":      chatID,
		"text":         text,
		"reply_markup": InlineKeyboard(eventID, stats),
	}

	var resp SendMessageResponse

	err := c.post(ctx, "sendMessage", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) post(ctx context.Context, method string, payload any, out any) error {
	url := fmt.Sprintf(
		"https://api.telegram.org/bot%s/%s",
		c.token,
		method,
	)

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	//nolint:gosec
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed sending request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if out == nil {
		return nil
	}

	if err = json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("telegram: failed to decode response: %w", err)
	}

	return nil
}
