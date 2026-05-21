package main

import (
	"context"
	"payment/internal/handler"
	"payment/internal/repository/postgres"
	"payment/internal/repository/redis"
	"payment/internal/service"
	"payment/pkg/cache"
	"payment/pkg/config"
	"payment/pkg/db"
	"payment/pkg/kafka"
	"payment/pkg/logger"
	"sync"

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

	cacheDB, err := cache.New(cfg.Cache, log)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := cacheDB.Close()
		if err != nil {
			panic(err)
		}
	}()
	log.Info("Successfully connected to cache")

	producer, err := kafka.NewProducer(cfg.Kafka)
	if err != nil {
		panic(err)
	}
	log.Info("Successfully connected to kafka")
	paymentProducerService := service.NewPaymentProducerImpl(cfg.Kafka, producer, log)
	paymentRepo := postgres.NewPaymentRepositoryImpl(database)
	outboxRepo := postgres.NewOutboxRepositoryImpl(database)
	transactionRepo := postgres.NewTransactionalRepositoryImpl(database, log)
	cacheRepository := redis.NewCacheRepository(cacheDB)
	paymentSvc := service.NewPaymentServiceImpl(paymentRepo, outboxRepo, cacheRepository, transactionRepo, log)
	workerPool := service.NewWorkerPoolImpl(cfg.Worker, paymentProducerService, outboxRepo, transactionRepo, log)
	paymentHandler := handler.NewPaymentHandlerImpl(paymentSvc, log)
	paymentRouter := handler.NewPaymentRouterImpl(paymentHandler)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		wg.Wait()
	}()

	workerPool.Start(ctx, &wg)

	paymentRouter.Route(cfg.Server)
}
