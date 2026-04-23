package main

import (
	"payment/internal/handler"
	"payment/internal/repository"
	"payment/internal/service"
)

func main() {

	paymentRepo := repository.NewPaymentRepositorySqlImpl()
	paymentSvc := service.NewPaymentServiceImpl(paymentRepo)
	var paymentHandler handler.PaymentHandler = handler.NewPaymentHandlerImpl(paymentSvc)
	var paymentRouter handler.PaymentRouter = handler.NewPaymentRouterImpl(paymentHandler)

	paymentRouter.Route()
}
