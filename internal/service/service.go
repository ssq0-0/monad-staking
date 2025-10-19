package service

import (
	"context"
	"crypto/ecdsa"
	"log"
	"ms/internal/models"
	"ms/pkg/utils"
	"sync"
	"time"
)

type (
	Client interface {
		SendTransaction(ctx context.Context, amount float32, to string, privatekey *ecdsa.PrivateKey, validatorID uint8) error
	}
)

type staker struct {
	monadClient Client
	ctx         context.Context
	wg          sync.WaitGroup
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

func (s *staker) Wait() {
	s.wg.Wait()
}

func (s *staker) Start(ctx context.Context, cfg RunParams, accounts []models.Account) {
	log.Printf("[INFO] Starting $MON staking process for %d accounts...", len(accounts))

	for i, acc := range accounts {
		select {
		case <-ctx.Done():
			log.Printf("[INFO] Context cancelled, stopping at account %d/%d", i, len(accounts))
			return
		default:
		}

		rndStake := utils.RanndomAmount(cfg.Stake.Min, cfg.Stake.Max)
		validatorID := utils.RandomSliceValue(cfg.Validators)

		s.wg.Add(1)
		go func(cfg RunParams, acc models.Account, stake float32, validator uint8) {
			defer s.wg.Done()

			select {
			case <-ctx.Done():
				log.Printf("[INFO] Context cancelled, skipping transaction for %s", acc.Address.Hex()[:10])
				return
			default:
			}

			if err := s.monadClient.SendTransaction(ctx, stake, cfg.ContractAddress, acc.PrivateKey, validator); err != nil {
				log.Printf("[WARN] failed stake for %s: %v", acc.Address.Hex()[:10], err)
			} else {
				log.Printf("[INFO] successfully staked %.4f MON for %s (validator: %d)", stake, acc.Address.Hex()[:10], validator)
			}
		}(cfg, acc, rndStake, validatorID)

		if i < len(accounts)-1 {
			rndSleep := utils.RanndomAmount(cfg.Delay.Min, cfg.Delay.Max)
			log.Printf("[INFO] waiting %.2f seconds before next account...", rndSleep)

			select {
			case <-time.After(time.Duration(rndSleep) * time.Second):
			case <-ctx.Done():
				log.Printf("[INFO] Context cancelled during delay, stopping...")
				return
			}
		}
	}
}
