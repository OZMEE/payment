package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"payment/internal/appers"
	"payment/internal/model"
	"payment/internal/service"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type PaymentHandler interface {
	GetAllPayments(w http.ResponseWriter, r *http.Request)
	GetPaymentById(w http.ResponseWriter, r *http.Request)
	PostPayment(w http.ResponseWriter, r *http.Request)
	PutPayment(w http.ResponseWriter, r *http.Request)
	DeletePayment(w http.ResponseWriter, r *http.Request)
}

type PaymentHandlerImpl struct {
	service service.PaymentService
}

func NewPaymentHandlerImpl(service service.PaymentService) *PaymentHandlerImpl {
	return &PaymentHandlerImpl{service: service}
}

func (h *PaymentHandlerImpl) GetAllPayments(w http.ResponseWriter, r *http.Request) {
	fmt.Println("getAllPayments request started")
	payments, err := h.service.GetAllPayments(r.Context())
	if err != nil {
		errorHandling(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(payments); err != nil {
		errorHandling(w, err)
		return
	}
}

func (h *PaymentHandlerImpl) GetPaymentById(w http.ResponseWriter, r *http.Request) {
	fmt.Println("getPaymentById request started")
	id, err := getId(r)
	if err != nil {
		errorHandling(w, err)
		return
	}

	payment, err := h.service.GetPaymentById(r.Context(), id)
	if err != nil {
		errorHandling(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(payment); err != nil {
		errorHandling(w, err)
		return
	}
}

func (h *PaymentHandlerImpl) PostPayment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("postPayment request started")
	dto, err := getPaymentDto(r)
	if err != nil {
		errorHandling(w, err)
		return
	}

	payment, err := h.service.PostPayment(r.Context(), dto)
	if err != nil {
		errorHandling(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(payment); err != nil {
		errorHandling(w, err)
		return
	}
}

func (h *PaymentHandlerImpl) PutPayment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("putPayment request started")
	id, err := getId(r)
	if err != nil {
		errorHandling(w, err)
		return
	}

	dto, err := getPaymentDto(r)
	if err != nil {
		errorHandling(w, err)
		return
	}

	if err := h.service.PutPayment(r.Context(), dto, id); err != nil {
		errorHandling(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *PaymentHandlerImpl) DeletePayment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("deletePayment request started")
	id, err := getId(r)
	if err != nil {
		errorHandling(w, err)
		return
	}

	if err := h.service.DeletePayment(r.Context(), id); err != nil {
		errorHandling(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func errorHandling(w http.ResponseWriter, err error) {
	var errResp *appers.ErrorResp
	if !errors.As(err, &errResp) {
		http.Error(w, fmt.Sprintf("%v (err = %s)", appers.ErrUnknown, err), appers.ErrUnknown.Code)
		return
	}
	http.Error(w, errResp.Error(), errResp.Code)
}

func getId(r *http.Request) (int64, error) {
	idString := chi.URLParam(r, "id")
	if idString == "" {
		return -1, appers.NewErrValidation("Payment ID is required")
	}
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		return -1, appers.NewErrValidation("Invalid format for Payment ID")
	}
	return id, nil
}

func getPaymentDto(r *http.Request) (*model.PaymentDto, error) {
	var dto model.PaymentDto
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		return nil, appers.NewErrValidation("Invalid format dto")
	}
	if dto.Amount <= 0 {
		return nil, appers.NewErrValidation("Invalid amount")
	}
	if dto.PaymentId <= 0 {
		return nil, appers.NewErrValidation("Invalid payment id")
	}
	return &dto, nil
}
