package retrieval

/*
	Calls the embedding_service (a small Python/FastAPI wrapper around the
	SAME model used at ingest time, all-MiniLM-L6-v2) so query-time vectors
	are comparable to what's already stored in the embeddings table. Go has
	no native equivalent to sentence-transformers, so this stays a small
	HTTP call rather than being reimplemented in Go.
*/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/pgvector/pgvector-go"
)

type embedRequest struct {
	Text string `json:"text"`
}

type embedResponse struct {
	Embedding []float32 `json:"embedding"`
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

func embeddingServiceURL() string {
	if v := os.Getenv("EMBEDDING_SERVICE_URL"); v != "" {
		return v
	}
	return "http://localhost:8001"
}

// EmbedQuery sends free text to the embedding microservice and returns a
// pgvector.Vector ready to use directly as a query parameter in search.go.

func EmbedQuery(text string) (pgvector.Vector, error) {
	body, err := json.Marshal(embedRequest{Text: text})
	if err != nil {
		return pgvector.Vector{}, fmt.Errorf("marshaling request: %w", err)
	}

	resp, err := httpClient.Post(
		embeddingServiceURL()+"/embed",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return pgvector.Vector{}, fmt.Errorf("calling embedding service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return pgvector.Vector{}, fmt.Errorf("embedding service returned %d: %s", resp.StatusCode, respBody)
	}

	var parsed embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return pgvector.Vector{}, fmt.Errorf("decoding response: %w", err)
	}

	return pgvector.NewVector(parsed.Embedding), nil
}
