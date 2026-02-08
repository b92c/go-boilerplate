#!/usr/bin/env bash
set -euo pipefail

REGION="${AWS_REGION:-us-east-1}"
ARTIFACT="/opt/serverless/health.zip"

log() { echo "[$(date +'%F %T')] $*"; }

# S3 bucket
log "Creating S3 bucket: s3://go-boilerplate-bucket"
awslocal s3 mb s3://go-boilerplate-bucket >/dev/null 2>&1 || true

# DynamoDB
log "Creating DynamoDB table: example-items"
awslocal dynamodb create-table \
  --table-name example-items \
  --attribute-definitions AttributeName=id,AttributeType=S \
  --key-schema AttributeName=id,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region "${REGION}" >/dev/null 2>&1 || true
table (para CRUD e healthcheck perceber presença do serviço)
# SQS queue
log "Creating SQS queue: example-queue"
QUEUE_URL=$(awslocal sqs create-queue --queue-name example-queue --query QueueUrl --output text 2>/dev/null || true)

# SNS topic
log "Creating SNS topic: example-topic"
TOPIC_ARN=$(awslocal sns create-topic --name example-topic --query TopicArn --output text 2>/dev/null || true)

# Subscript SQS on SNS
if [[ -n "${TOPIC_ARN}" ]]; then
  if [[ -z "${QUEUE_URL:-}" ]]; then
    QUEUE_URL=$(awslocal sqs get-queue-url --queue-name example-queue --query QueueUrl --output text 2>/dev/null || true)
  fi
  if [[ -n "${QUEUE_URL}" ]]; then
    QUEUE_ARN=$(awslocal sqs get-queue-attributes --queue-url "${QUEUE_URL}" --attribute-names QueueArn --query 'Attributes.QueueArn' --output text 2>/dev/null || true)
    if [[ -n "${QUEUE_ARN}" ]]; then
      log "Subscribing SQS to SNS"
      awslocal sns subscribe --topic-arn "${TOPIC_ARN}" --protocol sqs --notification-endpoint "${QUEUE_ARN}" >/dev/null 2>&1 || true
      POLICY=$(cat <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "Allow-SNS-SendMessage",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "sqs:SendMessage",
      "Resource": "${QUEUE_ARN}",
      "Condition": {"ArnEquals": {"aws:SourceArn": "${TOPIC_ARN}"}}
    }
  ]
}
EOF
)
      awslocal sqs set-queue-attributes --queue-url "${QUEUE_URL}" --attributes Policy="$(echo "$POLICY" | tr -d '\n')" >/dev/null 2>&1 || true
    fi
  fi
fi

# Lambda function
if [[ -f "${ARTIFACT}" ]]; then
  log "Deploying Go Lambda: health"
  if ! awslocal lambda get-function --function-name health >/dev/null 2>&1; then
    awslocal lambda create-function \
      --function-name health \
      --runtime provided.al2 \
      --role arn:aws:iam::000000000000:role/lambda-role \
      --handler bootstrap \
      --zip-file fileb://"${ARTIFACT}" \
      --timeout 10 --memory-size 128 >/dev/null 2>&1 || true
  else
    awslocal lambda update-function-code \
      --function-name health \
      --zip-file fileb://"${ARTIFACT}" >/dev/null 2>&1 || true
  fi

else
  log "Lambda artifact not found at ${ARTIFACT} (lambda-pack may have failed)."
fi

log "Bootstrap completed."
