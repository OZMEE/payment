package redis

import (
	"context"
	"encoding/json"
	"errors"
	"payment/internal/appers"
	"payment/internal/model"
	"payment/pkg/cache"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type CacheRepository interface {
	SetPayment(ctx context.Context, payment *model.Payment) error
	GetPayment(ctx context.Context, id int64) (*model.Payment, error)
	DelPayment(ctx context.Context, id int64) error
}

type CacheRepositoryImpl struct {
	db *cache.CacheDatabase
}

func NewCacheRepository(db *cache.CacheDatabase) *CacheRepositoryImpl {
	return &CacheRepositoryImpl{db: db}
}

func (c *CacheRepositoryImpl) SetPayment(ctx context.Context, payment *model.Payment) error {
	const op = "CacheRepositoryImpl.SetPayment"

	paymentJSON, err := json.Marshal(payment)
	if err != nil {
		return err
	}

	err = c.db.Set(ctx, strconv.FormatInt(payment.ID, 10), paymentJSON)
	if err != nil {
		return appers.ErrUpdateCacheData.Builder().Msg(err.Error()).Op(op).Build()
	}
	return nil
}

func (c *CacheRepositoryImpl) GetPayment(ctx context.Context, id int64) (*model.Payment, error) {
	const op = "CacheRepositoryImpl.GetPayment"

	val, err := c.db.Get(ctx, strconv.FormatInt(id, 10))
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var payment model.Payment
	err = json.Unmarshal([]byte(val), &payment)
	if err != nil {
		return nil, appers.ErrParseJson.Builder().Msg(err.Error()).Op(op).Build()
	}

	return &payment, nil
}

func (c *CacheRepositoryImpl) DelPayment(ctx context.Context, id int64) error {
	const op = "CacheRepositoryImpl.GetPayment"

	err := c.db.Del(ctx, strconv.FormatInt(id, 10))
	if err != nil {
		return appers.ErrUpdateCacheData.Builder().Msg(err.Error()).Op(op).Build()
	}

	return nil
}
