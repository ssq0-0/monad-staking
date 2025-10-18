package main

import (
	"context"
	"log"
	"ms/internal/client"
	"ms/internal/config"
	"ms/internal/models"
	"ms/internal/service"
	"path/filepath"
	"runtime"
)

func main() {
	ctx := context.Background()
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	configPath := filepath.Join(basepath, "..", "config.yaml")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ethClient, err := client.NewEthClient(ctx, cfg.RPCString)
	if err != nil {
		log.Fatalf("failed to init eth client: %v", err)
	}

	accounts, err := models.LoadAccountsFromFile(cfg.PrivateKeysFile)
	if err != nil {
		log.Fatalf("failed to init accounts: %v", err)
	}

	srv := service.NewStaker(ctx, ethClient)
	srv.Start(ctx, cfg, accounts)
}
