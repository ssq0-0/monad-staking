package client

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

func (c *EthClient) CallCA(toCA common.Address, data []byte) ([]byte, error) {
	callMsg := ethereum.CallMsg{
		To:   &toCA,
		Data: data,
	}

	return c.client.CallContract(context.Background(), callMsg, nil)
}

func (c *EthClient) GetNonce(address common.Address) uint64 {
	nonce, err := c.client.PendingNonceAt(context.Background(), address)
	if err != nil {
		return 0
	}
	return nonce
}

func (c *EthClient) GetChainID() (int64, error) {
	chainID, err := c.client.NetworkID(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to get ChainID: %w", err)
	}
	return chainID.Int64(), nil
}

func (c *EthClient) GetGasValues(msg ethereum.CallMsg) (uint64, *big.Int, *big.Int, error) {
	header, err := c.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("ошибка получения заголовка блока: %w", err)
	}

	maxPriorityFeePerGas, err := c.client.SuggestGasTipCap(context.Background())
	if err != nil {
		return 0, nil, nil, fmt.Errorf("ошибка получения предложения Gas Tip Cap: %w", err)
	}

	maxFeePerGas := new(big.Int).Add(header.BaseFee, maxPriorityFeePerGas)

	gasLimit, err := c.client.EstimateGas(context.Background(), msg)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("ошибка оценки газа: %w", err)
	}

	return gasLimit, maxPriorityFeePerGas, maxFeePerGas, nil
}
