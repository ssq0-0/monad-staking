package client

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type ChainData struct {
	Amount, ChainID, MaxPriorityFeePerGas, MaxFeePerGas *big.Int
	GasLimit, Nonce                                     uint64
	TxData                                              []byte
	DestinationAddr                                     *common.Address
}
