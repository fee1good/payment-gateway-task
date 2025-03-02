.PHONY: mocks
mocks:
	@echo "Generating mocks..."
	@mkdir -p tests/mocks
	@mockgen -source=internal/services/gateway_service.go -destination=tests/mocks/mock_gateway_service.go -package=mocks
	@mockgen -source=db/db_helpers.go -destination=tests/mocks/mock_storage.go -package=mocks
	@mockgen -source=internal/cache/redis.go -destination=tests/mocks/mock_cache.go -package=mocks
	@mockgen -source=internal/workers/transaction_processor.go -destination=tests/mocks/mock_transaction_processor.go -package=mocks
	@mockgen -source=internal/gateway/client.go -destination=tests/mocks/mock_gateway_client.go -package=mocks
	@mockgen -source=internal/kafka/producer.go -destination=tests/mocks/mock_kafka_producer.go -package=mocks
	@echo "Mocks generated successfully!"

# You can also add a dependency to ensure mocks are generated before tests
.PHONY: test
test: mocks
	go test ./... -v

# Add to the all target if you have one
.PHONY: all
all: mocks build test