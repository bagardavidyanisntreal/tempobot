package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	token string
}

func NewClient(token string) *Client {

	return &Client{token: token}
}

func (c *Client) EditMessage(
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

	req := map[string]interface{}{
		"chat_id":      chatID,
		"message_id":   messageID,
		"text":         text,
		"reply_markup": InlineKeyboard(eventID, stats),
	}

	body, _ := json.Marshal(req)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return err
}
