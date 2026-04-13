package main

import (
	"payment/internal/handler"
	"payment/internal/repository"
	"payment/internal/service"
)

func main() {

	var paymentRepoSql repository.PaymentRepositorySql = repository.NewPaymentRepositorySqlImpl()
	var paymentRepo repository.PaymentRepository = repository.NewPaymentRepositoryImpl(paymentRepoSql)
	var paymentSvc service.PaymentService = service.NewPaymentServiceImpl(paymentRepo)
	var paymentHandler handler.PaymentHandler = handler.NewPaymentHandlerImpl(paymentSvc)
	var paymentRouter handler.PaymentRouter = handler.NewPaymentRouterImpl(paymentHandler)

	paymentRouter.Route()
}
