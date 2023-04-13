package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Boyuan-Chen/v3-migration/flags"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli"
)

type Config struct {
	L2PrivateEndpoint string
	L2PublicEndpoint  string
	L2LegacyEndpoint  string
	JWTSecretPath     string
	MaxWaitingTime    int
	EpochLengthSecond int
	BobaHardForkBlock int
}

func NewConfig(ctx *cli.Context) *Config {
	cfg := Config{}
	cfg.L2PrivateEndpoint = ctx.GlobalString(flags.L2PrivateEndpointFlag.Name)
	cfg.L2PublicEndpoint = ctx.GlobalString(flags.L2PublicEndpointFlag.Name)
	cfg.MaxWaitingTime = ctx.GlobalInt(flags.MaxWaitingTimeFlag.Name)
	cfg.EpochLengthSecond = ctx.GlobalInt(flags.EpochLengthSecondFlag.Name)

	if ctx.GlobalIsSet(flags.L2LegacyEndpointFlag.Name) {
		cfg.L2LegacyEndpoint = ctx.GlobalString(flags.L2LegacyEndpointFlag.Name)
	} else {
		log.Crit("L2 Legacy Endpoint is not set")
	}

	if ctx.GlobalIsSet(flags.JWTSecretPathFlag.Name) {
		cfg.JWTSecretPath = ctx.GlobalString(flags.JWTSecretPathFlag.Name)
	} else {
		log.Crit("JWT Secret Path is not set")
	}

	if ctx.GlobalIsSet(flags.BobaHardForkBlockFlag.Name) {
		cfg.BobaHardForkBlock = ctx.GlobalInt(flags.BobaHardForkBlockFlag.Name)
	} else {
		log.Crit("Boba Hard Fork Block is not set")
	}

	return &cfg
}

func (cfg *Config) GetJWTSecret() (*[32]byte, error) {
	if cfg.JWTSecretPath == "" {
		return nil, errors.New("Failed to load file. The file location is empty.")
	}
	var secret [32]byte
	if data, err := os.ReadFile(cfg.JWTSecretPath); err == nil {
		jwtSecret := common.FromHex(strings.TrimSpace(string(data)))
		if len(jwtSecret) != 32 {
			return nil, fmt.Errorf("invalid jwt secret in path %s, not 32 hex-formatted bytes", cfg.JWTSecretPath)
		}
		copy(secret[:], jwtSecret)
		return &secret, nil
	} else {
		return nil, fmt.Errorf("Failed to read file: %s", err.Error())
	}
}
