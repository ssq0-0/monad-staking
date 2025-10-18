package service

import (
	"context"
	"crypto/ecdsa"
	"log"
	"ms/internal/config"
	"ms/internal/models"
	"ms/pkg/utils"
	"time"
)

type (
	Client interface {
		SendTransaction(ctx context.Context, amount float32, to string, privatekey *ecdsa.PrivateKey) error
	}
)

type staker struct {
	monadClient Client
	ctx         context.Context
}

func NewStaker(
	ctx context.Context,
	monadClient Client,
) *staker {
	return &staker{
		monadClient: monadClient,
		ctx:         ctx,
	}
}

func (s *staker) Close() {
	return
}

func (s *staker) Start(ctx context.Context, cfg *config.AppConfig, accounts []models.Account) {
	log.Println("[INFO] Starting $MON staking process...")

	for i := range accounts {
		rndStake := utils.RanndomAmount(cfg.Stake.Min, cfg.Stake.Max)
		validatorAddr := utils.RandomSliceValue(cfg.Validators)

		go func(cfg *config.AppConfig, acc models.Account) {
			if err := s.monadClient.SendTransaction(ctx, rndStake, validatorAddr, acc.PrivateKey); err != nil {
				log.Printf("[WARN] failed stake: %w", err)
			}
			log.Printf("[INFO] succesfully $MON staking [%s]", acc.Address[:4])
		}(cfg, accounts[i])

		rndSleep := utils.RanndomAmount(cfg.Delay.Min, cfg.Delay.Max)
		select {
		case <-time.After(time.Duration(rndSleep)):
		case <-ctx.Done():
			log.Fatalf("context cancel: %v", ctx.Err().Error())
		}
	}
}
