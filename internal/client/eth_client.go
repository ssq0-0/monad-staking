package client

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log"
	"math/big"
	client "ms/internal/client/consts"

	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/sync/errgroup"
)

var GlobalETHClient map[string]*EthClient

type EthClient struct {
	Client *ethclient.Client
}

func EthClientFactory(rpcs map[string]string) error {
	if len(rpcs) == 0 {
		return errors.New("RPC URLs map is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var (
		result = make(map[string]*EthClient)
		mu     sync.Mutex
		g, _   = errgroup.WithContext(ctx)
	)

	for name, rpc := range rpcs {
		name, rpc := name, rpc
		g.Go(func() error {
			client, err := ethclient.DialContext(ctx, rpc)
			if err != nil {
				return fmt.Errorf("error connecting to RPC %s: %v", name, err)
			}
			mu.Lock()
			result[name] = &EthClient{Client: client}
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	GlobalETHClient = result

	return nil
}

func CloseAllClients(clients map[string]*EthClient) {
	for _, client := range clients {
		if client.Client != nil {
			client.Client.Close()
		}
	}
}

func (c *EthClient) BalanceCheck(owner, tokenAddr common.Address) (*big.Int, error) {
	if isNative(tokenAddr) {
		balance, err := c.Client.BalanceAt(context.Background(), owner, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get native coin balance: %v", err)
		}
		return balance, nil
	}

	data, err := client.Erc20ABI.Pack("balanceOf", owner)
	if err != nil {
		return nil, fmt.Errorf("failed to pack data: %v", err)
	}

	result, err := c.CallCA(tokenAddr, data)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %v", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("empty response from ERC-20 contract")
	}

	var balance *big.Int
	if err := client.Erc20ABI.UnpackIntoInterface(&balance, "balanceOf", result); err != nil {
		return nil, fmt.Errorf("failed to unpack result: %v", err)
	}

	return balance, nil
}

func (c *EthClient) GetDecimals(token common.Address) (uint8, error) {
	if isNative(token) {
		return 18, nil
	}

	data, err := client.Erc20ABI.Pack("decimals")
	if err != nil {
		return 0, err
	}

	result, err := c.CallCA(token, data)
	if err != nil {
		return 0, fmt.Errorf("failed to call contract: %v", err)
	}

	var decimals uint8
	if err := client.Erc20ABI.UnpackIntoInterface(&decimals, "decimals", result); err != nil {
		return 0, fmt.Errorf("failed to unpack result: %v", err)
	}

	return decimals, nil
}

func (c *EthClient) CallCA(toCA common.Address, data []byte) ([]byte, error) {
	callMsg := ethereum.CallMsg{
		To:   &toCA,
		Data: data,
	}

	return c.Client.CallContract(context.Background(), callMsg, nil)
}

func (c *EthClient) GetGasValues(msg ethereum.CallMsg) (uint64, *big.Int, *big.Int, error) {
	timeout := time.After(time.Duration(5) * time.Minute)
	ticker := time.NewTicker(time.Second * time.Duration(5))
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			// log.Errorf("Gas wait timeout has been exceeded. Cycle interrupted.")
			return 0, big.NewInt(0), big.NewInt(0), fmt.Errorf("gas wait timeout has been exceeded, сycle interrupted")

		case <-ticker.C:
			header, err := c.Client.HeaderByNumber(context.Background(), nil)
			if err != nil {
				// log.Errorf("Ошибка получения заголовка блока: %v", err)
				return 0, nil, nil, fmt.Errorf("ошибка получения заголовка блока: %w", err)
			}

			maxPriorityFeePerGas, err := c.Client.SuggestGasTipCap(context.Background())
			if err != nil {
				// log.Errorf("Ошибка получения предложения Gas Tip Cap: %v", err)
				return 0, nil, nil, fmt.Errorf("ошибка получения предложения Gas Tip Cap: %w", err)
			}

			maxFeePerGas := new(big.Int).Add(header.BaseFee, maxPriorityFeePerGas)

			gasLimit, err := c.Client.EstimateGas(context.Background(), msg)
			if err != nil {
				// log.Errorf("Ошибка оценки газа: %v", err)
				return 0, nil, nil, fmt.Errorf("ошибка оценки газа: %w", err)
			}

			if maxFeePerGas.Cmp(big.NewInt(3000000000)) > 0 {
				log.Printf("[ATTENTION] High gwei %v", maxFeePerGas)
				continue
			} else {
				return gasLimit, maxPriorityFeePerGas, maxFeePerGas, nil
			}
		}
	}
}

func (c *EthClient) GetNonce(address common.Address) uint64 {
	nonce, err := c.Client.PendingNonceAt(context.Background(), address)
	if err != nil {
		return 0
	}
	return nonce
}

func (c *EthClient) GetChainID() (int64, error) {
	chainID, err := c.Client.NetworkID(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to get ChainID: %w", err)
	}
	return chainID.Int64(), nil
}

func (c *EthClient) ApproveTx(tokenAddr, spender common.Address, privateKey string, amount *big.Int) (*types.Transaction, error) {
	if isNative(tokenAddr) {
		return nil, nil
	}

	allowance, err := c.Allowance(tokenAddr, acc.Address, spender)
	if err != nil {
		return nil, fmt.Errorf("failed to get allowance: %v", err)
	}

	var approveValue *big.Int
	if allowance.Cmp(amount) >= 0 {
		return nil, nil
	}
	approveValue = client.MaxApproveValue

	approveData, err := client.Erc20ABI.Pack("approve", spender, approveValue)
	if err != nil {
		return nil, fmt.Errorf("failed to pack approve data: %v", err)
	}

	log.Printf("Approve transaction...")
	if err := c.SendTransaction(acc.PrivateKey, acc.Address, tokenAddr, c.GetNonce(acc.Address), big.NewInt(0), approveData); err != nil {
		return nil, err
	}

	time.Sleep(time.Second * 15)
	return nil, nil
}

func (c *EthClient) Allowance(tokenAddr, owner, spender common.Address) (*big.Int, error) {
	data, err := client.Erc20ABI.Pack("allowance", owner, spender)
	if err != nil {
		return nil, fmt.Errorf("failed to pack allowance data: %v", err)
	}

	msg := ethereum.CallMsg{
		To:   &tokenAddr,
		Data: data,
	}

	result, err := c.Client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %v", err)
	}

	var allowance *big.Int
	if err = client.Erc20ABI.UnpackIntoInterface(&allowance, "allowance", result); err != nil {
		return nil, fmt.Errorf("failed to unpack allowance data: %v", err)
	}

	return allowance, nil
}

func (c *EthClient) SendTransaction(privateKey *ecdsa.PrivateKey, ownerAddr, CA common.Address, value float32) error {
	chainID, err := c.Client.NetworkID(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get ChainID: %v", err)
	}

	gasLimit, maxPriorityFeePerGas, maxFeePerGas, err := c.GetGasValues(ethereum.CallMsg{
		From:  ownerAddr,
		To:    &CA,
		Value: value,
		Data:  txData,
	})
	if err != nil {
		return fmt.Errorf("failed to estimate gas: %v", err)
	}

	dynamicTx := types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: maxPriorityFeePerGas,
		GasFeeCap: maxFeePerGas,
		Gas:       gasLimit,
		To:        &CA,
		Value:     value,
		Data:      txData,
	}

	signedTx, err := types.SignTx(types.NewTx(&dynamicTx), types.LatestSignerForChainID(chainID), privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %v", err)
	}

	for attemps := 0; attemps < 5; attemps++ {
		if err = c.Client.SendTransaction(context.Background(), signedTx); err == nil {
			break
		} else {
			errorContext, critical := utils.IsCriticalError(err)
			if critical {
				return fmt.Errorf("failed to send transaction: %v", errorContext)
			}
		}
	}

	log.Printf("[NONCE: %v] Transaction sent: %s/tx/%s", nonce, client.ExploerLink[chainID.Int64()], signedTx.Hash().Hex())

	return c.waitForTransactionSuccess(signedTx.Hash(), 1*time.Minute)
}

func (c *EthClient) WaitTokenDeposit(chain, token string, accountAddress common.Address) error {
	chainId, err := c.GetChainID()
	if err != nil {
		return fmt.Errorf("ошибка получения chain id, проверка поступления средств отменена: %w", err)
	}

	tokenContract, ok := client.TokenContracts[chainId][token]
	if !ok {
		return fmt.Errorf("такого токена или сети в софте нет, остановка ожидания депозита на счет")
	}
	oldBalance, err := c.BalanceCheck(accountAddress, common.HexToAddress(tokenContract))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Cfg.DepositWaitingTime)*time.Minute)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("deposit waiting timout is closed")
		case <-ticker.C:
			currentBalance, err := c.BalanceCheck(accountAddress, common.HexToAddress(tokenContract))
			if err != nil {
				return fmt.Errorf("failed to check balance: %w", err)
			}

			if currentBalance.Cmp(oldBalance) == 0 {
				continue
			} else {
				return nil
			}
		}
	}
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
			receipt, err := c.Client.TransactionReceipt(context.Background(), txHash)
			if err != nil {
				if isUnknownBlockError(err) {
					continue
				}
				return fmt.Errorf("error getting transaction receipt: %v", err)
			}

			if receipt.Status == 1 {
				return nil
			} else {
				return fmt.Errorf("transaction failed")
			}
		}
	}
}

func isUnknownBlockError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "Unknown block") ||
		strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "free tier limits")
}

func isNative(token common.Address) bool {
	return token == common.Address{}
}
