package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/b92c/go-boilerplate/internal/usecase/health"
	"github.com/b92c/go-boilerplate/pkg/logger/zaplogger"

	"github.com/aws/aws-lambda-go/events"
)

// HealthResponse represents the response for the health check
// used both via local HTTP and API Gateway/Lambda
type HealthResponse struct {
	OK         bool   `json:"ok"`
	Message    string `json:"message"`
	LocalStack string `json:"localstackEndpoint"`
}

// LambdaHandler handles API Gateway (HTTP API) events
func LambdaHandler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Logger estruturado (DI leve aqui)
	logger, _ := zaplogger.FromEnv()
	defer logger.Sync()

	endpoint := os.Getenv("LOCALSTACK_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:4566"
	}

	healthSvc := health.NewServiceWithDeps(endpoint, nil, logger)
	resp := healthSvc.Check(ctx)

	b, _ := json.Marshal(resp)
	logger.Info("lambda health", "status", resp.OK, "message", resp.Message)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(b),
	}, nil
}

// HTTPHandler exposes /health for local execution without Lambda
func HTTPHandler(w http.ResponseWriter, r *http.Request) {
	logger, _ := zaplogger.FromEnv()
	defer logger.Sync()

	if r.URL.Path != "/health" {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		return
	}
	endpoint := os.Getenv("LOCALSTACK_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:4566"
	}

	healthSvc := health.NewServiceWithDeps(endpoint, nil, logger)
	resp := healthSvc.Check(r.Context())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
