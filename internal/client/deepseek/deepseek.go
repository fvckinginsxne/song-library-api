package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	apiClient"lyrics-library/internal/client"
)

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type Client struct {
	log    *slog.Logger
	client *http.Client
	token  string
}

const (
	messageContent = `You are a professional translator. Translate the following 
		song lyrics into Russian while preserving the poetic style, rhyme and rhythm 
		where possible. Keep the line breaks.`

	systemRole = "system"
	userRole   = "user"
	model      = "deepseek-chat"
	baseURL    = "https://api.deepseek.com/v1/chat/completions"
)

func New(log *slog.Logger, token string) *Client {
	return &Client{
		log:    log,
		client: &http.Client{},
		token:  token,
	}
}

func (c *Client) TranslateLyrics(ctx context.Context, lyrics []string) ([]string, error) {
	const op = "service.api.deepseek.TranslateLyrics"

	log := c.log.With(slog.String("op", op))

	log.Info("translating lyrics")

	ctx, cancel := context.WithTimeout(ctx, apiClient.RequestTimeout)
	defer cancel()

	req, err := c.buildAPIRequest(ctx, lyrics)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	res, err := c.doAPIRequest(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Debug("Deepseek response", slog.Any("response", res))

	if len(res.Choices) == 0 {
		return nil, apiClient.ErrFailedTranslateLyrics
	}

	formatted := apiClient.FormatLyrics(res.Choices[0].Message.Content)

	log.Info("lyrics translated successfully")

	return formatted, nil
}

func (c *Client) buildAPIRequest(ctx context.Context, lyrics []string) (*http.Request, error) {
	systemMessage := Message{
		Role:    systemRole,
		Content: messageContent,
	}

	userMessage := Message{
		Role:    userRole,
		Content: strings.Join(lyrics, "\n"),
	}

	requestData := ChatRequest{
		Model: model,
		Messages: []Message{
			systemMessage,
			userMessage,
		},
	}

	reqBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		baseURL,
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	return req, nil
}

func (c *Client) doAPIRequest(req *http.Request) (*ChatResponse, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response ChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
