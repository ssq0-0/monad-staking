package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadConfig загружает конфигурацию из YAML файла
func LoadConfig(configPath string) (*AppConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла конфигурации: %w", err)
	}

	var config AppConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга YAML: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("ошибка валидации конфигурации: %w", err)
	}

	return &config, nil
}

// validateConfig проверяет корректность конфигурации
func validateConfig(config *AppConfig) error {
	if config.Stake.Min < 0 {
		return fmt.Errorf("минимальное значение stake не может быть отрицательным")
	}
	if config.Stake.Max <= config.Stake.Min {
		return fmt.Errorf("максимальное значение stake должно быть больше минимального")
	}

	if config.Delay.Min < 0 {
		return fmt.Errorf("минимальное значение delay не может быть отрицательным")
	}
	if config.Delay.Max <= config.Delay.Min {
		return fmt.Errorf("максимальное значение delay должно быть больше минимального")
	}

	if len(config.Validators) == 0 {
		return fmt.Errorf("список валидаторов не может быть пустым")
	}

	if config.PrivateKeysFile == "" {
		return fmt.Errorf("путь к файлу с приватными ключами не может быть пустым")
	}

	return nil
}
