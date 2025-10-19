package client

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	client "ms/internal/client/consts"
	"ms/pkg/utils"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

func (c *EthClient) SendTransaction(ctx context.Context, amount float32, to string, privatekey *ecdsa.PrivateKey, validatorID uint8) error {
	preparedData, err := c.prepareData(ctx, amount, to, validatorID, privatekey)
	if err != nil {
		return fmt.Errorf("failed to prepare data: %v", err)
	}

	dynamicTx := types.DynamicFeeTx{
		ChainID:   preparedData.ChainID,
		Nonce:     preparedData.Nonce,
		GasTipCap: preparedData.MaxPriorityFeePerGas,
		GasFeeCap: preparedData.MaxFeePerGas,
		Gas:       preparedData.GasLimit,
		To:        preparedData.DestinationAddr,
		Value:     preparedData.Amount,
		Data:      preparedData.TxData,
	}

	signedTx, err := types.SignTx(types.NewTx(&dynamicTx), types.LatestSignerForChainID(preparedData.ChainID), privatekey)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %v", err)
	}

	for attempt := 0; attempt < client.RetryCount; attempt++ {
		if err = c.client.SendTransaction(context.Background(), signedTx); err == nil {
			break
		}
		log.Printf("[WARN] attempt %d failed: %v", attempt+1, err)
		time.Sleep(time.Second * 2)
	}

	log.Printf("[NONCE: %v] Transaction sent: %s/%s", preparedData.Nonce, client.ExploerTx, signedTx.Hash().Hex())

	return c.waitForTransactionSuccess(signedTx.Hash(), client.WaitingTimeout)
}

func (c *EthClient) prepareData(ctx context.Context, amount float32, to string, validatorID uint8, privatekey *ecdsa.PrivateKey) (ChainData, error) {
	bigAmount, err := utils.ConvertToWei(float64(amount), client.EthDecimal)
	if err != nil {
		return ChainData{}, err
	}

	chainID, err := c.client.NetworkID(ctx)
	if err != nil {
		return ChainData{}, fmt.Errorf("failed to get ChainID: %v", err)
	}

	ownerAddr, err := utils.DeriveAddress(privatekey)
	if err != nil {
		return ChainData{}, err
	}

	{
		balance, err := c.BalanceCheck(common.HexToAddress(to))
		if err != nil {
			return ChainData{}, err
		}

		if balance.Cmp(bigAmount) == -1 {
			return ChainData{}, fmt.Errorf("[%s] low balance: %v", ownerAddr, balance)
		}
	}

	txData, err := c.CreateDelegateData(validatorID)
	if err != nil {
		return ChainData{}, fmt.Errorf("failed to create delegate data: %v", err)
	}

	contract := common.HexToAddress(to)
	gasLimit, maxPriorityFeePerGas, maxFeePerGas, err := c.GetGasValues(ethereum.CallMsg{
		From:  ownerAddr,
		To:    &contract,
		Value: bigAmount,
		Data:  txData,
	})
	if err != nil {
		return ChainData{}, fmt.Errorf("failed to estimate gas: %v", err)
	}

	return ChainData{
		Amount:               bigAmount,
		ChainID:              chainID,
		MaxPriorityFeePerGas: maxPriorityFeePerGas,
		MaxFeePerGas:         maxFeePerGas,
		GasLimit:             gasLimit,
		Nonce:                c.GetNonce(ownerAddr),
		TxData:               txData,
		DestinationAddr:      &contract,
	}, nil
}

func (c *EthClient) waitForTransactionSuccess(txHash common.Hash, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("transaction wait timeout")
		case <-ticker.C:
			receipt, err := c.client.TransactionReceipt(context.Background(), txHash)
			if err != nil {
				log.Printf("error getting transaction receipt: %v", err)
				continue
			}

			if receipt.Status == 1 {
				return nil
			} else {
				return fmt.Errorf("transaction failed")
			}
		}
	}
}

func (c *EthClient) CreateDelegateData(validatorID uint8) ([]byte, error) {
	dataHex := client.DelegateSelector + fmt.Sprintf("%064x", validatorID)

	data, err := hexutil.Decode("0x" + dataHex)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания данных транзакции: %w", err)
	}

	return data, nil
}
