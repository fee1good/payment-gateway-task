# Project Notes

## To start the application

```
docker compose up
```
that's it

### curl examples

```shell
# Deposit transaction
curl -X POST http://localhost:8080/deposit \
  -H "Content-Type: application/json" \
  -d '{"user_id": 123, "amount": 100.50, "currency": "USD", "country": "US", "payment_method": "card"}'

# Withdrawal transaction
curl -X POST http://localhost:8080/withdrawal \
  -H "Content-Type: application/json" \
  -d '{"user_id": 123, "amount": 50.25, "currency": "EUR", "country": "DE", "payment_method": "bank_transfer"}'

# Gateway callback
curl -X POST http://localhost:8080/callback/42 \
  -H "Content-Type: application/json" \
  -d '{"gateway_transaction_id": "gw-tx-123", "status": "completed", "gateway_id": 1}'

````

## To-Do
- Use Goose for database migrations
- Use Swagger to generate API documentation
   - API request/response models will be moved from models to api package
- Use pgx or sqlx instead of default PostgreSQL driver
- Convert typed types to enums (e.g., transaction type, status, etc.)
- Load configuration values from Vault or similar secret management
- Don't use global logger instance; pass it as a parameter
   - Currently using global instance to simplify development
- Use existing retry libraries or embed retry logic into client
   - Avoid explicit/separate retry implementation
- Implement unmask mechanism in consumer services
   - The unmask method was deleted as we don't consume data from Kafka
- Review TODOs in the code


## Architectural Decisions

### Transaction Processing Flow

1. **Request Handling**: The API endpoints (`/deposit` and `/withdrawal`) receive transaction requests in either JSON or XML format.

2. **Validation**: Requests are validated for required fields and proper formatting.

3. **Transaction Creation**: A new transaction record is created in the database with "pending" status.

4. **Asynchronous Processing**: Transaction processing is handled asynchronously using a worker pool pattern to ensure scalability.

5. **Gateway Selection**: The system selects appropriate payment gateways based on the user's country, trying them in priority order.

6. **Callback Handling**: Gateway callbacks are processed asynchronously to update transaction status.

### Gateway Configuration and Selection

The system implements a region-based gateway selection mechanism:

1. **Country-Gateway Mapping**: Each country is mapped to one or more payment gateways in the database.

2. **Priority Order**: Gateways are tried in the order they are returned from the database, implementing an implicit priority system.

3. **Fallback Mechanism**: If a gateway fails to process a transaction, the system automatically tries the next available gateway for that country.

4. **Default Handling**: If no country-specific gateways are found, the system falls back to a default gateway.

### Fault Tolerance

1. **Circuit Breakers**: Implemented to prevent cascading failures when a gateway is consistently failing.

2. **Retry Mechanism**: Failed transactions are retried with exponential backoff.

3. **Transaction Locking**: Database transactions use row-level locking to prevent race conditions.

4. **Status Transition Protection**: Transactions in final states ("completed" or "failed") cannot be updated.

<details>
  <summary>--- App logs</summary>

```log
{"time":"2025-03-02T01:46:45.418638839Z","level":"INFO","msg":"Starting payment gateway service"}
2025-03-02T01:46:45.421395503Z {"time":"2025-03-02T01:46:45.420900647Z","level":"INFO","msg":"Successfully connected to the database."}
2025-03-02T01:46:45.424172040Z {"time":"2025-03-02T01:46:45.423570363Z","level":"INFO","msg":"Successfully connected to Redis","addr":"redis:6379"}
2025-03-02T01:46:45.425931242Z {"time":"2025-03-02T01:46:45.424527372Z","level":"INFO","msg":"Server starting","port":"8080"}
2025-03-02T01:47:20.002980082Z {"time":"2025-03-02T01:47:20.0002967Z","level":"INFO","msg":"Attempting to process payment with gateway","txID":1,"gatewayID":1,"attempt":"gateway-txn-1"}
2025-03-02T01:47:20.003026076Z {"time":"2025-03-02T01:47:20.000340695Z","level":"INFO","msg":"Successfully processed payment with gateway","txID":1,"gatewayID":1,"gatewayTxnID":"gateway-txn-1"}
2025-03-02T01:47:20.025918453Z {"time":"2025-03-02T01:47:20.025772554Z","level":"INFO","msg":"Message successfully published to Kafka","topic":"payment-transactions"}
### after sigterm
{"time":"2025-03-02T10:18:08.742703712Z","level":"INFO","msg":"Shutting down server..."}
2025-03-02T10:18:08.742876839Z {"time":"2025-03-02T10:18:08.742778796Z","level":"INFO","msg":"Shutting down HTTP server..."}
2025-03-02T10:18:08.744282682Z {"time":"2025-03-02T10:18:08.744216598Z","level":"INFO","msg":"Stopping transaction processor..."}
2025-03-02T10:18:08.744577392Z {"time":"2025-03-02T10:18:08.744354724Z","level":"INFO","msg":"Closing Kafka producer..."}
2025-03-02T10:18:08.744603392Z {"time":"2025-03-02T10:18:08.744411682Z","level":"INFO","msg":"Closing Redis connection..."}
2025-03-02T10:18:08.744624309Z {"time":"2025-03-02T10:18:08.744603184Z","level":"INFO","msg":"Closing database connection..."}
2025-03-02T10:18:08.745999610Z {"time":"2025-03-02T10:18:08.745923026Z","level":"INFO","msg":"Server gracefully stopped"}
```
</details>

<details>
  <summary>--- Test execution</summary>

```log
GOROOT=/opt/homebrew/opt/go/libexec #gosetup
GOPATH=/Users/viktorkorsunov/go #gosetup
/opt/homebrew/opt/go/libexec/bin/go test -c -o /Users/viktorkorsunov/Library/Caches/JetBrains/GoLand2024.3/tmp/GoLand/___1go_test_payment_gateway_internal_tests.test payment-gateway/internal/tests #gosetup
/opt/homebrew/opt/go/libexec/bin/go tool test2json -t /Users/viktorkorsunov/Library/Caches/JetBrains/GoLand2024.3/tmp/GoLand/___1go_test_payment_gateway_internal_tests.test -test.v=test2json -test.paniconexit0 #gosetup
=== RUN   TestHandleCallback_Success
--- PASS: TestHandleCallback_Success (0.00s)
=== RUN   TestHandleCallback_TransactionNotFound
--- PASS: TestHandleCallback_TransactionNotFound (0.00s)
=== RUN   TestHandleCallback_FailedStatus
--- PASS: TestHandleCallback_FailedStatus (0.00s)
=== RUN   TestHandleCallback_AlreadyInFinalState
--- PASS: TestHandleCallback_AlreadyInFinalState (0.00s)
=== RUN   TestHandleCallback_UpdateTransactionError
--- PASS: TestHandleCallback_UpdateTransactionError (0.00s)
=== RUN   TestHandleCallback_WithFallbackGateway
--- PASS: TestHandleCallback_WithFallbackGateway (0.00s)
=== RUN   TestHandleCallback_WithCustomStatus
--- PASS: TestHandleCallback_WithCustomStatus (0.00s)
=== RUN   TestHandleCallback_WithRejectedStatus
--- PASS: TestHandleCallback_WithRejectedStatus (0.00s)
=== RUN   TestProcessTransaction_Success
--- PASS: TestProcessTransaction_Success (0.00s)
=== RUN   TestProcessTransaction_CreateTransactionError
--- PASS: TestProcessTransaction_CreateTransactionError (0.00s)
=== RUN   TestProcessTransaction_KafkaError
--- PASS: TestProcessTransaction_KafkaError (0.00s)
=== RUN   TestProcessTransaction_MarshalError
--- PASS: TestProcessTransaction_MarshalError (0.00s)
=== RUN   TestProcessTransaction_WithSpecificGateway
--- PASS: TestProcessTransaction_WithSpecificGateway (0.00s)
=== RUN   TestProcessTransaction_WithdrawalWithPrioritizedGateways
--- PASS: TestProcessTransaction_WithdrawalWithPrioritizedGateways (0.00s)
=== RUN   TestProcessTransaction_WithHighValueAmount
--- PASS: TestProcessTransaction_WithHighValueAmount (0.00s)
=== RUN   TestGetTransactionStatus_Success
--- PASS: TestGetTransactionStatus_Success (0.00s)
=== RUN   TestGetTransactionStatus_NotFound
--- PASS: TestGetTransactionStatus_NotFound (0.00s)
=== RUN   TestGetTransactionStatus_WithGatewayInfo
--- PASS: TestGetTransactionStatus_WithGatewayInfo (0.00s)
=== RUN   TestGetTransactionStatus_FailedTransaction
--- PASS: TestGetTransactionStatus_FailedTransaction (0.00s)
=== RUN   TestGetTransactionStatus_DatabaseError
--- PASS: TestGetTransactionStatus_DatabaseError (0.00s)
PASS

Process finished with the exit code 0
```
</details>

###### The task took about 3 hours