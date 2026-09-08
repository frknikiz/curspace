package litellm

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestFetchModels(t *testing.T) {
	oldClient := modelHTTPClient
	defer func() { modelHTTPClient = oldClient }()
	modelHTTPClient = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/models" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Errorf("authorization header missing")
		}
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(strings.NewReader(`{"data":[{"id":"claude-sonnet","owned_by":"anthropic"},{"id":"claude-sonnet"},{"id":""}]}`)), Header: make(http.Header)}, nil
	})

	models, err := Fetch(context.Background(), "https://proxy.example.test", "secret")
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(models) != 1 || models[0].ID != "claude-sonnet" {
		t.Fatalf("models = %#v", models)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) Do(r *http.Request) (*http.Response, error) { return f(r) }

func TestModelsEndpointKeepsV1(t *testing.T) {
	endpoint, err := modelsEndpoint("https://example.test/v1/")
	if err != nil {
		t.Fatal(err)
	}
	if endpoint != "https://example.test/v1/models" {
		t.Fatalf("endpoint = %q", endpoint)
	}
}
