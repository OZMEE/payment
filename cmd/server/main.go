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
	database, err := db.New(cfg.Database)
	if err != nil {
		panic(err)
	}
	defer func(database *db.Database) {
		err := database.Close()
		if err != nil {
			panic(err)
		}
	}(database)

	log, err := logger.New()
	if err != nil {
		panic(err)
	}
	paymentRepo := repository.NewPaymentRepositoryImpl(database)
	paymentSvc := service.NewPaymentServiceImpl(paymentRepo)
	paymentHandler := handler.NewPaymentHandlerImpl(paymentSvc, log)
	paymentRouter := handler.NewPaymentRouterImpl(paymentHandler)

	paymentRouter.Route(cfg.Server)
}
