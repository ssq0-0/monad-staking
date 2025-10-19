package main

import (
	"context"
	"log"
	"ms/internal/client"
	"ms/internal/config"
	"ms/internal/models"
	"ms/internal/service"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

func main() {
	// Создаем контекст с возможностью отмены
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Настраиваем graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Горутина для обработки сигналов
	go func() {
		sig := <-sigChan
		log.Printf("[INFO] Получен сигнал %v, начинаем graceful shutdown...", sig)
		cancel() // Отменяем контекст
	}()

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

	srv.Start(ctx, service.RunParams{
		Stake:           service.Range{Min: cfg.Stake.Min, Max: cfg.Stake.Max},
		Delay:           service.Range{Min: cfg.Delay.Min, Max: cfg.Delay.Max},
		Validators:      cfg.Validators,
		ContractAddress: cfg.ContractAddress,
	}, accounts)

	log.Println("[INFO] Ожидание завершения всех транзакций...")

	done := make(chan struct{})
	go func() {
		srv.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("[INFO] Все транзакции завершены успешно.")
	case <-ctx.Done():
		log.Println("[INFO] Получен сигнал отмены, ожидаем завершения активных транзакций...")

		timeout := time.NewTimer(30 * time.Second)
		select {
		case <-done:
			log.Println("[INFO] Все активные транзакции завершены.")
		case <-timeout.C:
			log.Println("[WARN] Таймаут ожидания завершения транзакций, принудительное завершение.")
		}
	}

	log.Println("[INFO] Программа завершается.")
}
