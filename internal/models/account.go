package models

import (
	"crypto/ecdsa"
	"log"
	"ms/pkg/utils"

	"github.com/ethereum/go-ethereum/common"
)

type Account struct {
	Address    common.Address
	PrivateKey *ecdsa.PrivateKey
}

func (a *Account) CreateAccount(privateKey ...string) []Account {
	accounts := make([]Account, 0, len(privateKey))
	for _, pk := range privateKey {
		acc, err := processPrivateKeys(pk)
		if err != nil {
			log.Fatalf("failed to parse private keys: %w", err)
		}

		accounts = append(accounts, acc)
	}

	return accounts
}

func processPrivateKeys(input string) (Account, error) {
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
