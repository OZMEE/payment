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
	"go.uber.org/zap"
)

type PaymentHandler interface {
	GetAllPayments(w http.ResponseWriter, r *http.Request)
	GetPaymentById(w http.ResponseWriter, r *http.Request)
	PostPayment(w http.ResponseWriter, r *http.Request)
	PutPayment(w http.ResponseWriter, r *http.Request)
	DeletePayment(w http.ResponseWriter, r *http.Request)
}

type PaymentHandlerImpl struct {
	service         service.PaymentService
	producerService service.PaymentProducer
	log             *zap.Logger
}

func NewPaymentHandlerImpl(service service.PaymentService, producerService service.PaymentProducer, log *zap.Logger) *PaymentHandlerImpl {
	return &PaymentHandlerImpl{service: service, producerService: producerService, log: log}
}

func (h *PaymentHandlerImpl) GetAllPayments(w http.ResponseWriter, r *http.Request) {
	h.log.Info("getAllPayments request started",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path))

	payments, err := h.service.GetAllPayments(r.Context())
	if err != nil {
		h.errorHandling(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(payments); err != nil {
		h.errorHandling(w, err)
		return
	}
}

func (h *PaymentHandlerImpl) GetPaymentById(w http.ResponseWriter, r *http.Request) {
	h.log.Info("getPaymentById request started",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("id", chi.URLParam(r, "id")))

	id, err := h.getId(r)
	if err != nil {
		h.errorHandling(w, err)
		return
	}

	payment, err := h.service.GetPaymentById(r.Context(), id)
	if err != nil {
		h.errorHandling(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(payment); err != nil {
		h.errorHandling(w, err)
		return
	}
}

func (h *PaymentHandlerImpl) PostPayment(w http.ResponseWriter, r *http.Request) {
	h.log.Info("postPayment request started",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path))

	dto, err := h.getPaymentDto(r)
	if err != nil {
		h.errorHandling(w, err)
		return
	}

	payment, err := h.service.PostPayment(r.Context(), dto)
	if err != nil {
		h.errorHandling(w, err)
		return
	}

	if err := h.producerService.SendPaymentEvent(r.Context(), payment, "user-id"); err != nil {
		h.errorHandling(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(payment); err != nil {
		h.errorHandling(w, err)
		return
	}
}

func (h *PaymentHandlerImpl) PutPayment(w http.ResponseWriter, r *http.Request) {
	h.log.Info("putPayment request started",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path))

	id, err := h.getId(r)
	if err != nil {
		h.errorHandling(w, err)
		return
	}

	dto, err := h.getPaymentDto(r)
	if err != nil {
		h.errorHandling(w, err)
		return
	}

	if err := h.service.PutPayment(r.Context(), dto, id); err != nil {
		h.errorHandling(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *PaymentHandlerImpl) DeletePayment(w http.ResponseWriter, r *http.Request) {
	h.log.Info("deletePayment request started",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path))

	id, err := h.getId(r)
	if err != nil {
		h.errorHandling(w, err)
		return
	}

	if err := h.service.DeletePayment(r.Context(), id); err != nil {
		h.errorHandling(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *PaymentHandlerImpl) errorHandling(w http.ResponseWriter, err error) {
	var errResp *appers.ErrorResp
	if !errors.As(err, &errResp) {
		h.log.Error(err.Error())
		http.Error(w, fmt.Sprintf("%v (err = %s)", appers.ErrUnknown.SetMsg(err.Error()), err), appers.ErrUnknown.Code)
		return
	}

	if errResp.FieldName != "" {
		h.log.Error(errResp.Error(),
			zap.String(errResp.FieldName, errResp.FieldValue))
	} else {
		h.log.Error(errResp.Error())
	}

	http.Error(w, errResp.Error(), errResp.Code)
}

func (h *PaymentHandlerImpl) getId(r *http.Request) (int64, error) {
	idString := chi.URLParam(r, "id")
	if idString == "" {
		return -1, appers.NewErrValidation("Payment ID is required", "id", idString)
	}
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		return -1, appers.NewErrValidation("Invalid format for Payment ID", "id", idString)
	}
	return id, nil
}

func (h *PaymentHandlerImpl) getPaymentDto(r *http.Request) (*model.PaymentDto, error) {
	var dto model.PaymentDto
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.log.Error(err.Error())
		return nil, appers.ErrParseJson.SetMsg(err.Error())
	}
	if dto.Amount <= 0 {
		return nil, appers.NewErrValidation("Invalid amount", "amount", strconv.FormatInt(dto.Amount, 10))
	}
	if dto.PaymentId <= 0 {
		return nil, appers.NewErrValidation("Invalid payment_id", "payment_id", strconv.FormatInt(dto.PaymentId, 10))
	}
	return &dto, nil
}
