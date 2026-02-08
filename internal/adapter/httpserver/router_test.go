package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/b92c/go-boilerplate/internal/usecase/health"
)

type fakeHealthService struct{ resp health.Response }

func (f fakeHealthService) Check(_ context.Context) health.Response {
	return f.resp
}

func TestRouterHealthOK(t *testing.T) {
	r := NewRouter(fakeHealthService{resp: health.Response{OK: true, Message: "ok"}})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRouterHealthUnavailable(t *testing.T) {
	r := NewRouter(fakeHealthService{resp: health.Response{OK: false, Message: "down"}})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestRouterNotFound(t *testing.T) {
	r := NewRouter(fakeHealthService{resp: health.Response{OK: true}})
	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
