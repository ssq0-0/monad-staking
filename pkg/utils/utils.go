package utils

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand/v2"
	"regexp"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func ParsePrivateKey(hexKey string) (*ecdsa.PrivateKey, error) {
	if len(hexKey) != 64 && len(hexKey) != 66 {
		return nil, errors.New("invalid private key: incorrect length")
	}

	if len(hexKey) > 2 && hexKey[:2] == "0x" {
		hexKey = hexKey[2:]
	}

	isHex := regexp.MustCompile(`^[0-9a-fA-F]+$`).MatchString
	if !isHex(hexKey) {
		return nil, errors.New("invalid private key: contains non-hexadecimal characters")
	}

	privateKeyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	if len(privateKeyBytes) != 32 {
		return nil, errors.New("invalid private key format: incorrect byte length (must be 32 bytes)")
	}

	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	return privateKey, nil
}

func DeriveAddress(privateKey *ecdsa.PrivateKey) (common.Address, error) {
	if privateKey == nil {
		return (common.Address{}), fmt.Errorf("nil value in derive address function")
	}

	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	return crypto.PubkeyToAddress(*publicKey), nil
}

func RanndomAmount(min, max float32) float32 {
	return min + rand.Float32()*(max-min)
}

func RandomSliceValue[T any](slc []T) T {
	if len(slc) == 0 {
		panic("cannot pick random value from empty slice")
	}
	return slc[rand.IntN(len(slc))]
}
