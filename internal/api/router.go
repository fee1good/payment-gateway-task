package api

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"payment-gateway/db"
	"payment-gateway/internal/services"
)

func SetupRouter(dbHandler db.Storage, gatewayService services.GatewayServiceInterface) *mux.Router {
	router := mux.NewRouter()

	handler := NewTransactionHandler(dbHandler, gatewayService)

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := uuid.New().String()
			ctx := context.WithValue(r.Context(), "requestID", requestID)

			w.Header().Set("X-Request-ID", requestID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	router.HandleFunc("/deposit", handler.DepositHandler).Methods("POST")
	router.HandleFunc("/withdrawal", handler.WithdrawalHandler).Methods("POST")
	router.HandleFunc("/callback/{id:[0-9]+}", handler.CallbackHandler).Methods("POST")

	return router
}
