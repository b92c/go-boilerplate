package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/b92c/go-boilerplate/internal/usecase/health"
)

const (
	hContentType     = "Content-Type"
	vApplicationJSON = "application/json"
)

// ExampleService é a porta usada pelo adaptador HTTP para o CRUD de exemplo.
// Qualquer implementação com a mesma assinatura (como example.Service) atende.
type ExampleService interface {
	Create(ctx context.Context, item map[string]any) error
	Get(ctx context.Context, key map[string]any) (map[string]any, error)
	Update(ctx context.Context, item map[string]any) error
	Delete(ctx context.Context, key map[string]any) error
	List(ctx context.Context, limit int32) ([]map[string]any, error)
}

// Router representa o http.Handler da aplicação
// Segue princípios de Clean Architecture: camada de adapter (delivery)
// conhece apenas interfaces de caso de uso.
type Router struct {
	healthSvc  health.Service
	exampleSvc ExampleService // opcional
}

// NewRouter cria o roteador. exampleSvc é opcional para compatibilidade.
func NewRouter(healthSvc health.Service, exampleSvc ...ExampleService) http.Handler {
	r := &Router{healthSvc: healthSvc}
	if len(exampleSvc) > 0 {
		r.exampleSvc = exampleSvc[0]
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", r.healthHandler)
	mux.HandleFunc("/items", r.itemsHandler)
	mux.HandleFunc("/items/", r.itemHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(hContentType, vApplicationJSON)
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "route not found"})
	})
	return mux
}

func (r *Router) healthHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	resp := r.healthSvc.Check(req.Context())
	status := http.StatusOK
	if !resp.OK {
		status = http.StatusServiceUnavailable
	}
	w.Header().Set(hContentType, vApplicationJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

// /items
func (r *Router) itemsHandler(w http.ResponseWriter, req *http.Request) {
	if r.exampleSvc == nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "example service not configured"})
		return
	}
	switch req.Method {
	case http.MethodGet:
		limit := int32(50)
		if v := req.URL.Query().Get("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				limit = int32(n)
			}
		}
		items, err := r.exampleSvc.List(req.Context(), limit)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
			return
		}
		w.Header().Set(hContentType, vApplicationJSON)
		_ = json.NewEncoder(w).Encode(items)
	case http.MethodPost:
		var item map[string]any
		if err := json.NewDecoder(req.Body).Decode(&item); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid json"})
			return
		}
		if err := r.exampleSvc.Create(req.Context(), item); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
			return
		}
		w.Header().Set(hContentType, vApplicationJSON)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// /items/{id}
func (r *Router) itemHandler(w http.ResponseWriter, req *http.Request) {
	if r.exampleSvc == nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "example service not configured"})
		return
	}
	// extrai id após /items/
	id := strings.TrimPrefix(req.URL.Path, "/items/")
	if id == "" || strings.Contains(id, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	key := map[string]any{"id": id}

	switch req.Method {
	case http.MethodGet:
		item, err := r.exampleSvc.Get(req.Context(), key)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
			return
		}
		w.Header().Set(hContentType, vApplicationJSON)
		_ = json.NewEncoder(w).Encode(item)
	case http.MethodPut:
		var item map[string]any
		if err := json.NewDecoder(req.Body).Decode(&item); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid json"})
			return
		}
		// garante que id do path prevaleça
		item["id"] = id
		if err := r.exampleSvc.Update(req.Context(), item); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
			return
		}
		w.Header().Set(hContentType, vApplicationJSON)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	case http.MethodDelete:
		if err := r.exampleSvc.Delete(req.Context(), key); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
