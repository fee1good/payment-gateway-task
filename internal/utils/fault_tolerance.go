package utils

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/sony/gobreaker"
)

var cb = gobreaker.NewCircuitBreaker(gobreaker.Settings{
	Name:        "KafkaPublisher",
	MaxRequests: 1,
	Interval:    5 * time.Second,
	Timeout:     3 * time.Second,
	OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
		slog.Info("Circuit breaker state changed",
			"name", name,
			"from", from.String(),
			"to", to.String())
	},
})

func ExecuteWithCircuitBreaker(operation func() error) error {
	_, err := cb.Execute(func() (interface{}, error) {
		return nil, operation()
	})
	return err
}

func RetryOperation(operation func() error, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		if err := operation(); err == nil {
			return nil
		}
		time.Sleep(time.Duration(3600) * time.Second)
	}
	return fmt.Errorf("operation failed after %d attempts", maxRetries)
}
