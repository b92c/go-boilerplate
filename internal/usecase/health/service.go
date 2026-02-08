package health

import (
	"context"
	"net/http"
	"time"

	"github.com/b92c/go-boilerplate/internal/port"
)

type Response struct {
	OK         bool   `json:"ok"`
	Message    string `json:"message"`
	LocalStack string `json:"localstackEndpoint"`
}

type Service interface {
	Check(ctx context.Context) Response
}

func NewService(localstackEndpoint string) Service {
	return &service{localstackEndpoint: localstackEndpoint}
}

func NewServiceWithDeps(localstackEndpoint string, ddb port.DynamoDBPort, log port.Logger) Service {
	return &service{localstackEndpoint: localstackEndpoint, ddb: ddb, log: log}
}

type service struct {
	localstackEndpoint string
	ddb                port.DynamoDBPort
	log                port.Logger
}

func (s *service) Check(ctx context.Context) Response {
	ok := true
	msg := "api up"

	okLS, msgLS := s.checkLocalstack(ctx)
	if !okLS {
		ok = false
	}
	if msgLS != "" {
		msg = msg + "; " + msgLS
	}

	okDDB, msgDDB := s.checkDynamo(ctx)
	if !okDDB {
		ok = false
	}
	if msgDDB != "" {
		msg = msg + "; " + msgDDB
	}

	return Response{OK: ok, Message: msg, LocalStack: s.localstackEndpoint}
}

func (s *service) checkLocalstack(ctx context.Context) (bool, string) {
	if s.localstackEndpoint == "" {
		return true, ""
	}
	client := &http.Client{Timeout: 300 * time.Millisecond}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, s.localstackEndpoint+"/_localstack/health", nil)
	resp, err := client.Do(req)
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()
	if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if s.log != nil {
			s.log.Warn("localstack health failed", "error", err)
		}
		return false, "localstack unreachable"
	}
	return true, "localstack ok"
}

func (s *service) checkDynamo(ctx context.Context) (bool, string) {
	if s.ddb == nil {
		return true, ""
	}
	if err := s.ddb.Health(ctx); err != nil {
		if s.log != nil {
			s.log.Error("dynamodb health failed", "error", err)
		}
		return false, "dynamodb unreachable"
	}
	return true, "dynamodb ok"
}
