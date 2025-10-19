package client

import (
	"bytes"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

func SetInit() {
	parsedABI, err := abi.JSON(bytes.NewReader(Erc20JSON))
	if err != nil {
		log.Fatalf("Failed parsing ABI: %v", err)
	}

	Erc20ABI = &parsedABI
	_, success := MaxApproveValue.SetString("115792089237316195423570985008687907853269984665640564039457584007913129639935", 10)
	if !success {
		log.Fatalf("Failed to set MaxRepayBigInt: invalid number")
	}
}

// ####### Default GLOBALS #########
var (
	// Max approve value.
	MaxApproveValue = new(big.Int)

	// Default ABI for erc20 tokens
	Erc20ABI *abi.ABI

	EthDecimal = 18

	RetryCount = 5

	ExploerTx = "https://testnet.monadexplorer.com/tx"

	WaitingTimeout = 1 * time.Minute

	DelegateSelector = "84994fec"
)

// ###### Base ERC20 ABI. #######
// # Supplement other ABIs here #
var (
	Erc20JSON = []byte(`[
	{
		
		"constant":true,
		"inputs":[{"name":"account","type":"address"}],
		"name":"balanceOf",
		"outputs":[{"name":"","type":"uint256"}],
		"payable":false,
		"stateMutability":"view",
		"type":"function"
	},
	{
		"constant":true,
		"inputs":[{"name":"spender","type":"address"},{"name":"owner","type":"address"}],
		"name":"allowance",
		"outputs":[{"name":"","type":"uint256"}],
		"payable":false,
		"stateMutability":"view",
		"type":"function"
	},
	{
		"constant": false,
		"inputs": [],
		"name": "deposit",
		"outputs": [],
		"payable": true,
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
			"internalType": "uint256",
			"name": "wad",
			"type": "uint256"
			}
		],
		"name": "withdraw",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant":false,
		"inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],
		"name":"approve",
		"outputs":[{"name":"","type":"bool"}],
		"payable":false,
		"stateMutability":"nonpayable",
		"type":"function"
	},
	{
		"constant":false,
		"inputs":[{"name":"recipient","type":"address"},{"name":"amount","type":"uint256"}],
		"name":"transfer",
		"outputs":[{"name":"","type":"bool"}],
		"payable":false,
		"stateMutability":"nonpayable",
		"type":"function"
	},
	{
		"constant":false,
		"inputs":[{"name":"sender","type":"address"},{"name":"recipient","type":"address"},{"name":"amount","type":"uint256"}],
		"name":"transferFrom",
		"outputs":[{"name":"","type":"bool"}],
		"payable":false,
		"stateMutability":"nonpayable",
		"type":"function"
	},
	{
		"constant":true,
		"inputs":[],
		"name":"decimals",
		"outputs":[{"name":"","type":"uint8"}],
		"payable":false,
		"stateMutability":"view",
		"type":"function"
	},
	{
		"constant":true,
		"inputs":[],
		"name":"name",
		"outputs":[{"name":"","type":"string"}],
		"payable":false,
		"stateMutability":"view",
		"type":"function"
	},
	{
		"constant":true,
		"inputs":[],
		"name":"symbol",
		"outputs":[{"name":"","type":"string"}],
		"payable":false,
		"stateMutability":"view",
		"type":"function"
	},
	{
		"constant":true,
		"inputs":[],
		"name":"totalSupply",
		"outputs":[{"name":"","type":"uint256"}],
		"payable":false,
		"stateMutability":"view",
		"type":"function"
	}
]`)
)
