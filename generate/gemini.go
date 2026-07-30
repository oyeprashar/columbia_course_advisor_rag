package generate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// gemini.google.com's free tier currently covers Flash / Flash-Lite models
// (Pro models moved to paid-only in April 2026) -- check
// https://ai.google.dev/gemini-api/docs/pricing for the current free-tier
// model list, since this changes fairly often. Override with GEMINI_MODEL
// if the default below is no longer free-tier eligible.
var geminiAPIBaseURL = "https://generativelanguage.googleapis.com/v1beta/models"

func geminiModelName() string {
	if v := os.Getenv("GEMINI_MODEL"); v != "" {
		return v
	}
	return "gemini-2.5-flash"
}

type geminiRequest struct {
	SystemInstruction geminiContent   `json:"system_instruction"`
	Contents          []geminiContent `json:"contents"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"` // omitted for system_instruction
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func callGemini(systemPrompt, userMessage string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY not set")
	}

	reqBody := geminiRequest{
		SystemInstruction: geminiContent{
			Parts: []geminiPart{{Text: systemPrompt}},
		},
		Contents: []geminiContent{
			{Role: "user", Parts: []geminiPart{{Text: userMessage}}},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	url := fmt.Sprintf("%s/%s:generateContent?key=%s", geminiAPIBaseURL, geminiModelName(), apiKey)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling Gemini API: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	var parsed geminiResponse
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return "", fmt.Errorf("decoding response: %w (body: %s)", err, respBytes)
	}

	if parsed.Error != nil {
		return "", fmt.Errorf("Gemini API error: %s", parsed.Error.Message)
	}

	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from Gemini API (body: %s)", respBytes)
	}

	return parsed.Candidates[0].Content.Parts[0].Text, nil
}
