package main

import (
	"payment/internal/handler"
	"payment/internal/repository"
	"payment/internal/service"
	"payment/pkg/config"
	"payment/pkg/db"
	"payment/pkg/kafka"
	"payment/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	log, err := logger.New()
	if err != nil {
		panic(err)
	}
	log.Info("Logger initialized")

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log.Info("Configuration loaded",
		zap.Any("config", cfg))

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
	log.Info("Successfully connected to db")

	producer, err := kafka.NewProducer(cfg.Kafka)
	if err != nil {
		panic(err)
	}
	log.Info("Successfully connected to kafka")
	paymentProducerService := service.NewPaymentProducerImpl(producer, log)
	paymentRepo := repository.NewPaymentRepositoryImpl(database)
	paymentSvc := service.NewPaymentServiceImpl(paymentRepo)
	paymentHandler := handler.NewPaymentHandlerImpl(paymentSvc, paymentProducerService, log)
	paymentRouter := handler.NewPaymentRouterImpl(paymentHandler)

	paymentRouter.Route(cfg.Server)
}
