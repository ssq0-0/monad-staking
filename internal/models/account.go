package models

import (
	"bufio"
	"crypto/ecdsa"
	"fmt"
	"log"
	"ms/pkg/utils"
	"os"

	"github.com/ethereum/go-ethereum/common"
)

type Account struct {
	Address    common.Address
	PrivateKey *ecdsa.PrivateKey
}

// LoadAccountsFromFile загружает аккаунты из файла с приватными ключами
func LoadAccountsFromFile(filePath string) ([]Account, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла с приватными ключами: %w", err)
	}
	defer file.Close()

	var privateKeys []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		privateKeys = append(privateKeys, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ошибка чтения файла: %w", err)
	}

	if len(privateKeys) == 0 {
		return nil, fmt.Errorf("файл не содержит приватных ключей")
	}

	return CreateAccounts(privateKeys...), nil
}

func CreateAccounts(privateKeys ...string) []Account {
	accounts := make([]Account, 0, len(privateKeys))
	for _, pk := range privateKeys {
		acc, err := processPrivateKey(pk)
		if err != nil {
			log.Fatalf("failed to parse private key: %v", err)
		}

		accounts = append(accounts, acc)
	}

	return accounts
}

func processPrivateKey(input string) (Account, error) {
	priv, err := utils.ParsePrivateKey(input)
	if err != nil {
		return Account{}, err
	}

	addr, err := utils.DeriveAddress(priv)
	if err != nil {
		return Account{}, err
	}

	return Account{
		Address:    addr,
		PrivateKey: priv,
	}, nil
}
