package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum/go-ethereum/common"
)

type Config struct {
	// The path to the private key file.
	SecretConfigPath string `json:"secretPath"`

	// The path to the rollup config file.
	RollupConfigPath string `json:"rollupConfigPath"`
}

func NewConfig(secretPath string, rollupConfigPath string) *Config {
	return &Config{
		SecretConfigPath: secretPath,
		RollupConfigPath: rollupConfigPath,
	}
}

func (cfg *Config) GetSecret() (*[32]byte, error) {
	if cfg.SecretConfigPath == "" {
		return nil, errors.New("Failed to load file. The file location is empty.")
	}
	var secret [32]byte
	if data, err := os.ReadFile(cfg.SecretConfigPath); err == nil {
		jwtSecret := common.FromHex(strings.TrimSpace(string(data)))
		if len(jwtSecret) != 32 {
			return nil, fmt.Errorf("invalid jwt secret in path %s, not 32 hex-formatted bytes", cfg.SecretConfigPath)
		}
		copy(secret[:], jwtSecret)
		return &secret, nil
	} else {
		return nil, fmt.Errorf("Failed to read file: %s", err.Error())
	}
}

func (cfg *Config) GetRollupConfig() (*rollup.Config, error) {
	file, err := os.Open(cfg.RollupConfigPath)
	defer file.Close()
	if err != nil {
		return nil, errors.New("failed to read rollup config")
	}
	defer file.Close()
	var rollupConfig rollup.Config
	if err := json.NewDecoder(file).Decode(&rollupConfig); err != nil {
		return nil, fmt.Errorf("to decode rollup config: %s", err.Error())
	}
	return &rollupConfig, nil
}

func (cfg *Config) GetConfig() (*rollup.Config, *[32]byte, error) {
	// load JWT key
	secret, err := cfg.GetSecret()
	if err != nil {
		return nil, nil, err
	}
	// load rollup config
	rollupConfig, err := cfg.GetRollupConfig()
	if err != nil {
		return nil, nil, err
	}
	return rollupConfig, secret, nil
}
