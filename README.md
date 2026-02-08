# Go Boilerplate (Go + AWS Lambda + LocalStack Web + Serverless)

A clean architecture Go boilerplate ready for:

- Local development with Docker Engine (docker compose)
- AWS simulation via LocalStack + LocalStack UI
- Lambda/API Gateway (HTTP handler) and local binary (HTTP server)
- Native unit tests (without mock frameworks)
- Deployment with Serverless Framework (dev/test/prod)

## Directory Structure (based on Bob Martin's Clean Architecture, aka 'Uncle Bob')

```text
.
├── cmd/
│   ├── main.go                 # HTTP entrypoint for local development
│   └── lambda/main.go          # Lambda entrypoint (AWS)
├── internal/
│   ├── adapter/
│   │   ├── httpserver/
│   │   │   ├── router.go       # HTTP adapter (delivery)
│   │   │   └── router_test.go
│   │   └── repository/
│   │       └── dynamo/
│   │           └── example_repository.go  # CRUD repository using DynamoDB port
│   ├── handler/
│   │   └── lambda.go           # API Gateway handler (Lambda) with structured logs
│   ├── port/
│   │   ├── dynamodb.go         # Port (interface) for DynamoDB
│   │   └── logger.go           # Port (interface) for Logger
│   └── usecase/
│       ├── health/
│       │   ├── service.go      # Healthcheck use case (injects Logger/DynamoDB)
│       │   └── service_test.go
│       └── example/
│           └── service.go      # Example CRUD use case using DynamoDB repo
├── pkg/
│   ├── dynamodb/
│   │   └── client.go           # Implementação AWS SDK v2 (DynamoDBPort)
│   └── logger/
│       └── zaplogger/
│           └── zaplogger.go    # Implementação com zap (Logger)
├── .setup/localstack-web/      # Web UI for LocalStack
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── serverless.yml
├── events/
│   └── health-event.json
├── go.mod / go.sum
└── README.md
```

Architectural Decision

- Domain layer (usecase) depends only on ports (interfaces) in `internal/port`.
- Infrastructure implementations reside in `pkg/*` (e.g., AWS SDK v2, zap), isolated from the domain.
- Adapters expose I/O (HTTP) and repositories implement orchestration for use cases.
- Composition and DI occur at entrypoints (`cmd/*`).

### Ports and pkg (examples)

- Logger (port): `internal/port/logger.go`
  - Implementation: `pkg/logger/zaplogger`
- DynamoDB (port): `internal/port/dynamodb.go`
  - Implementation: `pkg/dynamodb/client.go`
  - Example repository: `internal/adapter/repository/dynamo/example_repository.go`

### Dependency Injection

- Local entrypoint (`cmd/main.go`):

```go
zl, _ := zaplogger.FromEnv()
ddbClient, _ := dynamodb.New(ctx, dynamodb.Options{Region: region, Endpoint: localstackEndpoint})
healthSvc := health.NewServiceWithDeps(localstackEndpoint, ddbClient, zl)
repo := dynamo.NewExampleRepository(ddbClient, "example-items")
exampleSvc := example.NewService(repo, zl)
router := httpserver.NewRouter(healthSvc, exampleSvc)
```

- Lambda entrypoint (`cmd/lambda/main.go` + `internal/handler/lambda.go`):

```go
lambda.Start(handler.LambdaHandler) // handler creates logger and injects into use case
```

## Local Development

1. Start the stack:

```bash
docker compose up -d --build
```

1. LocalStack Web UI:

- <http://localhost:8081>

### Interface Examples (LocalStack UI)

- #### Dashboard

  **Light** mode
  ![LocalStack UI - Light](.setup/docs/localstack-ui.png)

  **Dark** mode
  ![LocalStack UI - Dark](.setup/docs/localstack-ui-dark.png)

1. Healthcheck:

```bash
curl http://localhost:8080/health
```

1. Example CRUD (/items)

- Create the table in LocalStack (DynamoDB):
  - Name: `example-items`
  - Partition key: `id (String)`

- Create an item:

```bash
curl -s -X POST http://localhost:8080/items \
  -H 'Content-Type: application/json' \
  -d '{"id":"1","name":"foo"}'
```

- List items:

```bash
curl -s http://localhost:8080/items
```

- Get an item:

```bash
curl -s http://localhost:8080/items/1
```

- Update:

```bash
curl -s -X PUT http://localhost:8080/items/1 \
  -H 'Content-Type: application/json' \
  -d '{"name":"bar"}'
```

- Delete:

```bash
curl -s -X DELETE http://localhost:8080/items/1 -i
```

## Example Stack (LocalStack bootstrap)

When you start Docker (docker compose up -d --build), LocalStack automatically executes the `localstack/01-bootstrap.sh` script (init/ready.d) and creates resources for development:

- S3: bucket `s3://go-boilerplate-bucket`
- DynamoDB: table `example-items (PK: id String)`
- Lambda: function `health` (Go, runtime provided.al2)
- SNS: topic `example-topic`
- SQS: queue `example-queue` subscribed to the SNS topic

### Quick Examples (awslocal)

- Check created resources:

```bash
awslocal s3 ls
awslocal dynamodb list-tables
awslocal lambda list-functions
awslocal sns list-topics
awslocal sqs list-queues
```

- Invoke the example Lambda:

```bash
awslocal lambda invoke \
  --function-name health \
  --payload '{"rawPath":"/health","requestContext":{"http":{"method":"GET","path":"/health"}}}' \
  /tmp/out.json >/dev/null && cat /tmp/out.json; echo
```

- Publish to SNS and read from SQS:

```bash
TOPIC_ARN=$(awslocal sns list-topics --query 'Topics[0].TopicArn' --output text)
awslocal sns publish --topic-arn "$TOPIC_ARN" --message 'hello from sns'
QUEUE_URL=$(awslocal sqs get-queue-url --queue-name example-queue --query QueueUrl --output text)
awslocal sqs receive-message --queue-url "$QUEUE_URL" --wait-time-seconds 1
```

- S3: upload and list files:

```bash
awslocal s3 cp README.md s3://go-boilerplate-bucket/README.md
awslocal s3 ls s3://go-boilerplate-bucket
```

- DynamoDB: insert and list items:

```bash
awslocal dynamodb put-item \
  --table-name example-items \
  --item '{"id":{"S":"test-1"},"name":{"S":"foo"}}'
awslocal dynamodb scan --table-name example-items
```

## Tests

```bash
go test ./...
```

## Build and Package Lambda (custom runtime provided.al2)

```bash
make package-lambda
```

This generates `.serverless/health.zip` containing the `bootstrap` binary (required by the runtime).

## Deploy with Serverless

- Prerequisites: Node 18+, npm/npx.
- With LocalStack (development):

```bash
REGION=us-east-1 make sls-deploy
```

- Remove:

```bash
REGION=us-east-1 make sls-remove
```

- For real AWS, authenticate your AWS CLI/credentials and run the same command without LocalStack running.

## Environment Variables

- LOCALSTACK_ENDPOINT: container (<http://localstack:4566>), host (<http://localhost:4566>)
- AWS_REGION: us-east-1 (default)
- APP_ENV: dev/local/prod (defines logger preset)
- ENABLE_DYNAMODB: false to disable client (enabled by default)
- EXAMPLE_TABLE: table name for CRUD (default: example-items)

## Thanks & Credits

This project uses [LocalStack Web](https://github.com/dantasrafael/localstack-web), a web UI for LocalStack. Special thanks to [dantasrafael](https://github.com/dantasrafael) for creating and maintaining this excellent tool.

## License

MIT
