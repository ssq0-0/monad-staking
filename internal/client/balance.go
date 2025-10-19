package client

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func (c *EthClient) BalanceCheck(owner common.Address) (*big.Int, error) {
	balance, err := c.client.BalanceAt(context.Background(), owner, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get native coin balance: %v", err)
	}

	return balance, nil
}
