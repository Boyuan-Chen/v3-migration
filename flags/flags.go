package flags

import (
	"github.com/urfave/cli"
)

var (
	L2PrivateEndpointFlag = cli.StringFlag{
		Name:   "l2-private-endpoint",
		Value:  "http://localhost:8551",
		Usage:  "L2 Private Endpoint",
		EnvVar: "L2_PRIVATE_ENDPOINT",
	}
	L2PublicEndpointFlag = cli.StringFlag{
		Name:   "l2-public-endpoint",
		Value:  "http://127.0.0.1:9545",
		Usage:  "L2 Public Endpoint",
		EnvVar: "L2_PUBLIC_ENDPOINT",
	}
	L2LegacyEndpointFlag = cli.StringFlag{
		Name:   "l2-legacy-endpoint",
		Usage:  "L2 Legacy Endpoint",
		EnvVar: "L2_LEGACY_ENDPOINT",
	}
	JWTSecretPathFlag = cli.StringFlag{
		Name:   "jwt-secret-path",
		Usage:  "Path to JWT secret",
		EnvVar: "JWT_SECRET_PATH",
	}
	MaxWaitingTimeFlag = cli.IntFlag{
		Name:   "max-waiting-time",
		Value:  5,
		Usage:  "Maximum waiting time for a transaction to be mined (second)",
		EnvVar: "MAX_WAITING_TIME",
	}
	EpochLengthSecondFlag = cli.IntFlag{
		Name:   "epoch-length-second",
		Value:  1,
		Usage:  "Epoch length in second",
		EnvVar: "EPOCH_LENGTH_SECOND",
	}
)

var Flags = []cli.Flag{
	L2PrivateEndpointFlag,
	L2PublicEndpointFlag,
	L2LegacyEndpointFlag,
	JWTSecretPathFlag,
	MaxWaitingTimeFlag,
	EpochLengthSecondFlag,
}
