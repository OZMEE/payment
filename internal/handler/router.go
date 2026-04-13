package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type PaymentRouter interface {
	Route()
}

type PaymentRouterImpl struct {
	handler PaymentHandler
}

func NewPaymentRouterImpl(handler PaymentHandler) *PaymentRouterImpl {
	return &PaymentRouterImpl{handler: handler}
}

func (p PaymentRouterImpl) Route() {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(15 * time.Second))

	r.Route("/api", func(r chi.Router) {
		r.Get("/payments", p.handler.getAllPayments)
		r.Post("/payments", p.handler.postPayment)
		r.Get("/payments/{id}", p.handler.getPaymentById)
		r.Put("/payments/{id}", p.handler.putPayment)
		r.Delete("/payments/{id}", p.handler.deletePayment)
	})
	if err := http.ListenAndServe(":8080", r); err != nil {
		fmt.Printf("Server failed: %v\n", err)
		return
	}
}
