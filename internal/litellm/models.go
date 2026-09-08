package litellm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Model is a model exposed by a LiteLLM OpenAI-compatible proxy.
type Model struct {
	ID      string `json:"id"`
	OwnedBy string `json:"owned_by,omitempty"`
}

type modelsResponse struct {
	Data []Model `json:"data"`
}

var modelHTTPClient interface {
	Do(*http.Request) (*http.Response, error)
} = &http.Client{Timeout: 15 * time.Second}

// Fetch returns the models published by a LiteLLM proxy. Both a proxy root
// (https://proxy.example.com) and an OpenAI-style base URL
// (https://proxy.example.com/v1) are accepted.
func Fetch(ctx context.Context, baseURL, token string) ([]Model, error) {
	endpoint, err := modelsEndpoint(baseURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating LiteLLM model request: %w", err)
	}
	if strings.TrimSpace(token) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))
	}
	resp, err := modelHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching LiteLLM models: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("litellm model request returned %s", resp.Status)
	}

	var payload modelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decoding LiteLLM models: %w", err)
	}
	models := payload.Data[:0]
	seen := make(map[string]struct{}, len(payload.Data))
	for _, model := range payload.Data {
		model.ID = strings.TrimSpace(model.ID)
		if model.ID == "" {
			continue
		}
		if _, ok := seen[model.ID]; ok {
			continue
		}
		seen[model.ID] = struct{}{}
		models = append(models, model)
	}
	if len(models) == 0 {
		return nil, fmt.Errorf("litellm returned no models")
	}
	return models, nil
}

func modelsEndpoint(baseURL string) (string, error) {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return "", fmt.Errorf("litellm base URL is not configured")
	}
	u, err := url.Parse(baseURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid LiteLLM base URL %q", baseURL)
	}
	u.Path = strings.TrimRight(u.Path, "/")
	if !strings.HasSuffix(u.Path, "/v1") {
		u.Path += "/v1"
	}
	u.Path += "/models"
	return u.String(), nil
}
