package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dspm/internal/service"
	"dspm/internal/storage"
)

func TestHealthz(t *testing.T) {
	api := New(service.New(storage.NewMemoryRepository()))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	res := httptest.NewRecorder()

	api.Handler().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}
}

func TestCreateAgentAPI(t *testing.T) {
	api := New(service.New(storage.NewMemoryRepository()))
	body := bytes.NewBufferString(`{"name":"data-assistant","model":"gpt-4o-mini"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents", body)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	api.Handler().ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d, body=%s", res.Code, res.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload["error"] != nil {
		t.Fatalf("expected no error, got %#v", payload["error"])
	}
}
