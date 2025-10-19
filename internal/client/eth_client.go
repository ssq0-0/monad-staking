package client

import (
	"context"
	"errors"
	"fmt"

	"strings"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

type EthClient struct {
	client *ethclient.Client
}

func NewEthClient(ctx context.Context, rpc string) (*EthClient, error) {
	if strings.TrimSpace(rpc) == "" {
		return nil, errors.New("RPC is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, rpc)
	if err != nil {
		return nil, fmt.Errorf("error connecting to RPC %s: %v", rpc, err)
	}

	return &EthClient{
		client: client,
	}, nil
}
