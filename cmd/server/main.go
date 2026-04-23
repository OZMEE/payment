package main

import (
	"payment/internal/handler"
	"payment/internal/repository"
	"payment/internal/service"
	"payment/pkg/config"
	"payment/pkg/db"
	"payment/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	db, err := db.New(cfg.Database)
	if err != nil {
		panic(err)
	}
	log, err := logger.New()
	if err != nil {
		panic(err)
	}
	paymentRepo := repository.NewPaymentRepositoryImpl(db)
	paymentSvc := service.NewPaymentServiceImpl(paymentRepo)
	paymentHandler := handler.NewPaymentHandlerImpl(paymentSvc, *log)
	paymentRouter := handler.NewPaymentRouterImpl(paymentHandler)

	paymentRouter.Route()
}
