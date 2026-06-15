package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type GeminiClient struct {
	apiKey string
	model  string
}

func NewGeminiClient() *GeminiClient {
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.0-flash"
	}

	return &GeminiClient{
		apiKey: os.Getenv("GEMINI_API_KEY"),
		model:  model,
	}
}

func (c *GeminiClient) SendPrompt(systemPrompt, userPrompt string) (string, error) {
	if c.apiKey == "" {
		return "", errors.New("serviço de IA não configurado")
	}

	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		c.model, c.apiKey,
	)

	body := map[string]interface{}{
		"system_instruction": map[string]interface{}{
			"parts": []map[string]string{
				{"text": systemPrompt},
			},
		},
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": userPrompt},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("erro ao montar requisição: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao comunicar com a API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler resposta da API: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("erro ao processar otimização")
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("erro ao processar resposta da IA: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("erro ao processar otimização")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}
