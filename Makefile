APP_NAME=go-boilerplate
LAMBDA_NAME=health
BIN_DIR=bin
PKG=./...

.PHONY: build run test lint ci deps clean package-lambda

up:
	docker compose up -d --build

down:
	docker compose down

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BIN_DIR)/api ./cmd/main.go
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BIN_DIR)/bootstrap ./cmd/lambda/main.go

run:
	LOCALSTACK_ENDPOINT=http://localhost:4566 go run ./cmd/main.go

test:
	go test -race -cover $(PKG)

clean:
	rm -rf $(BIN_DIR) .serverless

package-lambda: build
	mkdir -p .serverless
	cd $(BIN_DIR) && zip -9 ../.serverless/$(LAMBDA_NAME).zip bootstrap

sls-deploy:
	npx serverless deploy --stage $${STAGE:-dev} --region $${REGION:-us-east-1}

sls-remove:
	npx serverless remove --stage $${STAGE:-dev} --region $${REGION:-us-east-1}

health:
	curl http://localhost:8080/health