package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"payment/internal/model"
	"payment/internal/service"

	"github.com/go-chi/chi/v5"
)

type PaymentHandler interface {
	getAllPayments(w http.ResponseWriter, r *http.Request)
	getPaymentById(w http.ResponseWriter, r *http.Request)
	postPayment(w http.ResponseWriter, r *http.Request)
	putPayment(w http.ResponseWriter, r *http.Request)
	deletePayment(w http.ResponseWriter, r *http.Request)
}

type PaymentHandlerImpl struct {
	service service.PaymentService
}

func NewPaymentHandlerImpl(service service.PaymentService) *PaymentHandlerImpl {
	return &PaymentHandlerImpl{service: service}
}

func (h PaymentHandlerImpl) getAllPayments(w http.ResponseWriter, r *http.Request) {
	fmt.Println("getAllPayments request started")
	payments, err := h.service.GetAllPayments(r.Context())
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(payments); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h PaymentHandlerImpl) getPaymentById(w http.ResponseWriter, r *http.Request) {
	fmt.Println("getPaymentById request started")
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Payment ID is required", http.StatusBadRequest)
		return
	}

	payment, err := h.service.GetPaymentById(r.Context(), id)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(payment); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h PaymentHandlerImpl) postPayment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("postPayment request started")
	var payment model.PaymentDto
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.service.PostPayment(r.Context(), &payment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	type response struct {
		ID string `json:"id"`
	}
	res := response{ID: id}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h PaymentHandlerImpl) putPayment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("putPayment request started")
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Payment ID is required", http.StatusBadRequest)
		return
	}

	var req model.PaymentDto
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.PutPayment(r.Context(), &req, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h PaymentHandlerImpl) deletePayment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("deletePayment request started")
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Payment ID is required", http.StatusBadRequest)
		return
	}

	if err := h.service.DeletePayment(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
