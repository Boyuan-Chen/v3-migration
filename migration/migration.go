package migration

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Boyuan-Chen/v3-migration/config"
	"github.com/Boyuan-Chen/v3-migration/engineapi"
	"github.com/Boyuan-Chen/v3-migration/mine"
	"github.com/Boyuan-Chen/v3-migration/rpc"
	"github.com/ethereum/go-ethereum/log"
)

var (
	defaultL2PrivateEndpoint = "http://localhost:8551"
	defaultL2PublicEndpoint  = "http://localhost:9545"

	errNoL2LegacyEndpoint = errors.New("no l2 legacy endpoint provided")
	errNoJWTSecretPath    = errors.New("no JWT secret path provided")
)

type Migration struct {
	config *config.Config
	ctx    context.Context
	stop   chan struct{}
	miner  *mine.Miner
}

func NewMigration(cfg *config.Config) (*Migration, error) {
	if cfg.L2PrivateEndpoint == defaultL2PrivateEndpoint {
		log.Info("L2 private endpoint is set to the default value.", "endpoint", defaultL2PrivateEndpoint)
	}
	if cfg.L2PublicEndpoint == defaultL2PublicEndpoint {
		log.Info("L2 public endpoint is set to the default value.", "endpoint", defaultL2PublicEndpoint)
	}
	if cfg.L2LegacyEndpoint == "" {
		return nil, fmt.Errorf("L2 legacy endpoint is not set: %w", errNoL2LegacyEndpoint)
	}
	if cfg.JWTSecretPath == "" {
		return nil, fmt.Errorf("JWT secret path is not set: %w", errNoJWTSecretPath)
	}

	JWTSecret, err := cfg.GetJWTSecret()
	if err != nil {
		return nil, err
	}
	l2PublicRpc, err := rpc.NewRpcClient(cfg.L2PublicEndpoint, *JWTSecret)
	if err != nil {
		return nil, err
	}
	l2LegacyRpc, err := rpc.NewRpcClient(cfg.L2LegacyEndpoint, *JWTSecret)
	if err != nil {
		return nil, err
	}
	l2PrivateRpc, err := rpc.NewRpcClient(cfg.L2PrivateEndpoint, *JWTSecret)
	if err != nil {
		return nil, err
	}
	l2EngineAPI, err := engineapi.NewEngineAPI(l2PrivateRpc, cfg)
	if err != nil {
		return nil, err
	}

	miner := mine.NewMiner(l2PublicRpc, l2LegacyRpc, l2EngineAPI, cfg)

	migration := &Migration{
		config: cfg,
		ctx:    context.Background(),
		stop:   make(chan struct{}),
		miner:  miner,
	}

	return migration, nil
}

func (m *Migration) Start() error {
	go m.Loop()
	return nil
}

func (m *Migration) Stop() {
	close(m.stop)
}

func (m *Migration) Wait() {
	<-m.stop
}

// Loop is the main logic of the migration
func (m *Migration) Loop() {
	timer := time.NewTicker(time.Duration(m.config.EpochLengthSecond) * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			log.Trace("polling", "time", time.Now())
			if err := m.miner.MineBlock(); err != nil {
				log.Error("cannot mine new block", "message", err)
			}
		case <-m.ctx.Done():
			m.Stop()
		}
	}
}

func (m *Migration) Mine() error {
	return nil
}
