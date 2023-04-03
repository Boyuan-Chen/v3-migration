package utils

import "github.com/ethereum/go-ethereum/common"

var (
	// L1StandardBridgeAddress is the address of the L1 standard bridge contract.
	L1StandardBridgeAddress = common.HexToAddress("0x6900000000000000000000000000000000000003")

	// L2StandardBridgeAddress is the address of the L2 standard bridge contract.
	L2StandardBridgeAddress = common.HexToAddress("0x4200000000000000000000000000000000000010")

	// L1MessengerAddress is the address of the L1 messenger contract.
	L1MessengerAddress = common.HexToAddress("0x6900000000000000000000000000000000000002")

	// L2MessengerAddress is the address of the L2 messenger contract.
	L2MessengerAddress = common.HexToAddress("0x4200000000000000000000000000000000000007")

	// L1BobaAddress is the address of the L1 Boba contract.
	L1BobaAddress = common.HexToAddress("0x154C5E3762FbB57427d6B03E7302BDA04C497226")

	// L2BobaAddress is the address of the L2 Boba contract.
	L2BobaAddress = common.HexToAddress("0x42000000000000000000000000000000000000fe")
)
